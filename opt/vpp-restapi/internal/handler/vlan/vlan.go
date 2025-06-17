package vlan

import (
    "fmt"
    "log"
    "net"
    "strings"

    "github.com/gin-gonic/gin"
    vppapi "vpp-restapi/internal/api"
    vppintf "vpp-restapi/binapi/interface"
    vppinterface_types "vpp-restapi/binapi/interface_types"
    vppip_types "vpp-restapi/binapi/ip_types"
)

// VLANCreateRequest is the request struct for creating a VLAN interface
type VLANCreateRequest struct {
    ParentIfIndex uint32 `json:"parent_if_index" binding:"required"`
    VLANID        uint32 `json:"vlan_id" binding:"required"`
    Enable        bool   `json:"enable,omitempty"`
    MTU           uint32 `json:"mtu,omitempty"`
    IPAddress     string `json:"ip_address,omitempty"`
}

// VLANActionRequest for enable/disable/mtu/ip
type VLANActionRequest struct {
    SwIfIndex uint32 `json:"sw_if_index" binding:"required"`
    MTU       uint32 `json:"mtu,omitempty"`
    IPAddress string `json:"ip_address,omitempty"`
}

// RegisterRoutes sets up the VLAN HTTP routes
// PATCH: gunakan gin.IRoutes dan daftarkan endpoint tanpa membuat group di sini
func RegisterRoutes(r gin.IRoutes, vppClient *vppapi.VPPClient) {
    r.POST("/vlan/create", createVLANHandler(vppClient))
    r.DELETE("/vlan/:sw_if_index", deleteVLANHandler(vppClient))
    r.POST("/vlan/:sw_if_index/enable", enableVLANHandler(vppClient, true))
    r.POST("/vlan/:sw_if_index/disable", enableVLANHandler(vppClient, false))
    r.POST("/vlan/:sw_if_index/mtu", setVLANMtuHandler(vppClient))
    r.POST("/vlan/:sw_if_index/ip", setVLANIpHandler(vppClient))
}

// @Summary Create VLAN Subinterface
// @Description Create a VLAN subinterface with options.
// @Tags vlan
// @Accept json
// @Produce json
// @Param body body VLANCreateRequest true "VLAN Config"
// @Success 201 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/vlan/create [post]
func createVLANHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req VLANCreateRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            log.Printf("VPP channel error: %v", err)
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        subifReq := &vppintf.CreateVlanSubif{
            SwIfIndex: vppinterface_types.InterfaceIndex(req.ParentIfIndex),
            VlanID:    req.VLANID,
        }
        subifReply := &vppintf.CreateVlanSubifReply{}
        if err := ch.SendRequest(subifReq).ReceiveReply(subifReply); err != nil || subifReply.Retval != 0 {
            c.JSON(500, gin.H{"error": "failed to create vlan", "details": err})
            return
        }

        vlanIfIndex := uint32(subifReply.SwIfIndex)
        // Enable if requested
        if req.Enable {
            enableReq := &vppintf.SwInterfaceSetFlags{
                SwIfIndex: vppinterface_types.InterfaceIndex(vlanIfIndex),
                Flags:     vppinterface_types.IF_STATUS_API_FLAG_ADMIN_UP,
            }
            enableReply := &vppintf.SwInterfaceSetFlagsReply{}
            _ = ch.SendRequest(enableReq).ReceiveReply(enableReply)
        }
        // Set MTU if requested
        if req.MTU > 0 {
            mtuReq := &vppintf.SwInterfaceSetMtu{
                SwIfIndex: vppinterface_types.InterfaceIndex(vlanIfIndex),
                Mtu:       []uint32{req.MTU, 0, 0, 0},
            }
            mtuReply := &vppintf.SwInterfaceSetMtuReply{}
            _ = ch.SendRequest(mtuReq).ReceiveReply(mtuReply)
        }
        // Set IP if requested
        if req.IPAddress != "" {
            ipReq := &vppintf.SwInterfaceAddDelAddress{
                SwIfIndex: vppinterface_types.InterfaceIndex(vlanIfIndex),
                IsAdd:     true,
                DelAll:    false,
                Prefix:    parseIPPrefix(req.IPAddress),
            }
            ipReply := &vppintf.SwInterfaceAddDelAddressReply{}
            _ = ch.SendRequest(ipReq).ReceiveReply(ipReply)
        }

        c.JSON(201, gin.H{
            "message":     "VLAN created",
            "sw_if_index": vlanIfIndex,
        })
    }
}

// @Summary Delete VLAN Subinterface
// @Description Delete a VLAN subinterface by sw_if_index.
// @Tags vlan
// @Param sw_if_index path int true "VLAN SwIfIndex"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/vlan/{sw_if_index} [delete]
func deleteVLANHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        swIfIndex := c.Param("sw_if_index")
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        delReq := &vppintf.DeleteSubif{
            SwIfIndex: parseSwIfIndex(swIfIndex),
        }
        delReply := &vppintf.DeleteSubifReply{}
        if err := ch.SendRequest(delReq).ReceiveReply(delReply); err != nil || delReply.Retval != 0 {
            c.JSON(500, gin.H{"error": "failed to delete vlan", "details": err})
            return
        }
        c.JSON(200, gin.H{"message": "VLAN deleted"})
    }
}

// @Summary Enable VLAN
// @Description Enable a VLAN interface by sw_if_index.
// @Tags vlan
// @Param sw_if_index path int true "VLAN SwIfIndex"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/vlan/{sw_if_index}/enable [post]
// @Router /vpp/vlan/{sw_if_index}/disable [post]
func enableVLANHandler(vppClient *vppapi.VPPClient, enable bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        swIfIndex := parseSwIfIndex(c.Param("sw_if_index"))
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        flags := vppinterface_types.IfStatusFlags(0)
        if enable {
            flags = vppinterface_types.IF_STATUS_API_FLAG_ADMIN_UP
        }
        req := &vppintf.SwInterfaceSetFlags{
            SwIfIndex: swIfIndex,
            Flags:     flags,
        }
        reply := &vppintf.SwInterfaceSetFlagsReply{}
        if err := ch.SendRequest(req).ReceiveReply(reply); err != nil || reply.Retval != 0 {
            c.JSON(500, gin.H{"error": "failed to set state", "details": err})
            return
        }
        status := "disabled"
        if enable {
            status = "enabled"
        }
        c.JSON(200, gin.H{"message": "VLAN " + status})
    }
}

// @Summary Set VLAN MTU
// @Description Set the MTU of a VLAN interface.
// @Tags vlan
// @Accept json
// @Produce json
// @Param sw_if_index path int true "VLAN SwIfIndex"
// @Param body body VLANActionRequest true "MTU Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/vlan/{sw_if_index}/mtu [post]
func setVLANMtuHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req VLANActionRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        mtuReq := &vppintf.SwInterfaceSetMtu{
            SwIfIndex: vppinterface_types.InterfaceIndex(req.SwIfIndex),
            Mtu:       []uint32{req.MTU, 0, 0, 0},
        }
        mtuReply := &vppintf.SwInterfaceSetMtuReply{}
        if err := ch.SendRequest(mtuReq).ReceiveReply(mtuReply); err != nil || mtuReply.Retval != 0 {
            c.JSON(500, gin.H{"error": "failed to set mtu", "details": err})
            return
        }
        c.JSON(200, gin.H{"message": "MTU set"})
    }
}

// @Summary Set VLAN IP Address
// @Description Set the IP address of a VLAN interface.
// @Tags vlan
// @Accept json
// @Produce json
// @Param sw_if_index path int true "VLAN SwIfIndex"
// @Param body body VLANActionRequest true "IP Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]interface{}
// @Router /vpp/vlan/{sw_if_index}/ip [post]
func setVLANIpHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req VLANActionRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        ipReq := &vppintf.SwInterfaceAddDelAddress{
            SwIfIndex: vppinterface_types.InterfaceIndex(req.SwIfIndex),
            IsAdd:     true,
            DelAll:    false,
            Prefix:    parseIPPrefix(req.IPAddress),
        }
        ipReply := &vppintf.SwInterfaceAddDelAddressReply{}
        if err := ch.SendRequest(ipReq).ReceiveReply(ipReply); err != nil || ipReply.Retval != 0 {
            c.JSON(500, gin.H{"error": "failed to set ip address", "details": err})
            return
        }
        c.JSON(200, gin.H{"message": "IP address set"})
    }
}

// Helper: parse string to swIfIndex
func parseSwIfIndex(str string) vppinterface_types.InterfaceIndex {
    var idx uint32
    _, _ = fmt.Sscanf(str, "%d", &idx)
    return vppinterface_types.InterfaceIndex(idx)
}

// Helper: parse CIDR ip string to Prefix (ip_types.AddressWithPrefix)
func parseIPPrefix(ip string) vppip_types.AddressWithPrefix {
    var addr vppip_types.Address
    var plen uint8
    ipstr, mask, found := strings.Cut(ip, "/")
    if !found {
        mask = "32"
    }
    parsed := net.ParseIP(ipstr)
    if parsed.To4() != nil {
        addr.Af = vppip_types.ADDRESS_IP4
        copy(addr.Un.XXX_UnionData[:], parsed.To4())
    } else {
        addr.Af = vppip_types.ADDRESS_IP6
        copy(addr.Un.XXX_UnionData[:], parsed.To16())
    }
    fmt.Sscanf(mask, "%d", &plen)
    return vppip_types.AddressWithPrefix{Address: addr, Len: plen}
}
