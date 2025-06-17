package lcpng

import (
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
    vppapi "vpp-restapi/internal/api"
    vppinterface_types "vpp-restapi/binapi/interface_types"
    vpp_lcp "vpp-restapi/binapi/lcpng_if"
)

// RegisterRoutes sets up the LCP-related HTTP routes.
// Sekarang menerima gin.IRoutes, tanpa membuat group di sini.
func RegisterRoutes(r gin.IRoutes, vppClient *vppapi.VPPClient) {
    r.POST("/lcp/mirror", mirrorInterfaceHandler(vppClient))
}

// @Summary Mirror VPP Interface to Host
// @Description Mirror a VPP interface (LCP pair) to the host's kernel namespace
// @Tags lcp
// @Accept json
// @Produce json
// @Param body body object true "Mirror Config {sw_if_index: int, host_if_name: string, host_if_type: string, netns: string}"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/lcp/mirror [post]
func mirrorInterfaceHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        var reqBody struct {
            SwIfIndex   uint32 `json:"sw_if_index" binding:"required"`
            HostIfName  string `json:"host_if_name,omitempty"`
            HostIfType  string `json:"host_if_type,omitempty"` // "tap" or "tun"
            Netns       string `json:"netns,omitempty"`
        }
        if err := c.ShouldBindJSON(&reqBody); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
            return
        }

        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            log.Printf("Failed to create API channel: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "channel creation failed", "details": err.Error()})
            return
        }
        defer ch.Close()

        // Set default namespace (if not set, default to "dataplane")
        netns := reqBody.Netns
        if netns == "" {
            netns = "dataplane"
        }
        nsSetReq := &vpp_lcp.LcpDefaultNsSet{Netns: netns}
        nsSetReply := &vpp_lcp.LcpDefaultNsSetReply{}
        if err := ch.SendRequest(nsSetReq).ReceiveReply(nsSetReply); err != nil {
            log.Printf("Failed to set default netns: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "set netns failed", "details": err.Error()})
            return
        }
        if nsSetReply.Retval != 0 {
            log.Printf("Set netns failed: retval=%d", nsSetReply.Retval)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "set netns failed", "details": "VPP returned non-zero retval"})
            return
        }

        // Mirror interface (Add LCP pair)
        var hostType vpp_lcp.LcpItfHostType = vpp_lcp.LCP_API_ITF_HOST_TAP
        if reqBody.HostIfType == "tun" {
            hostType = vpp_lcp.LCP_API_ITF_HOST_TUN
        }
        hostIfName := reqBody.HostIfName // optional

        pairReq := &vpp_lcp.LcpItfPairAddDel{
            IsAdd:      true,
            SwIfIndex:  vppinterface_types.InterfaceIndex(reqBody.SwIfIndex),
            HostIfName: hostIfName,
            HostIfType: hostType,
            Netns:      netns,
        }
        pairReply := &vpp_lcp.LcpItfPairAddDelReply{}
        if err := ch.SendRequest(pairReq).ReceiveReply(pairReply); err != nil {
            log.Printf("Failed to add LCP pair: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "add LCP pair failed", "details": err.Error()})
            return
        }
        if pairReply.Retval != 0 {
            log.Printf("Add LCP pair failed: retval=%d", pairReply.Retval)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "add LCP pair failed", "details": "VPP returned non-zero retval"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message":     "Interface mirrored to kernel namespace",
            "sw_if_index": reqBody.SwIfIndex,
            "netns":       netns,
        })
    }
}
