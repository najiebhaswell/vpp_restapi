package _interface

import (
    "github.com/gin-gonic/gin"
    "vpp-restapi/internal/api"
    vppintf "vpp-restapi/binapi/interface"
    vppintftypes "vpp-restapi/binapi/interface_types"
    "net/http"
)

func RegisterConfigRoutes(r gin.IRoutes, vppClient *api.VPPClient) {
    r.GET("/vpp/interfaces/config", getInterfacesConfigHandler(vppClient))
}

func getInterfacesConfigHandler(vppClient *api.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "channel creation failed"})
            return
        }
        defer ch.Close()

        req := &vppintf.SwInterfaceDump{}
        reply := &vppintf.SwInterfaceDetails{}
        reqCtx := ch.SendMultiRequest(req)

        interfaces := map[string]interface{}{}

        for {
            stop, err := reqCtx.ReceiveReply(reply)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed"})
                return
            }
            if stop {
                break
            }

            name := string(reply.InterfaceName)
            iface := map[string]interface{}{
                "description": "",
                "mac":         reply.L2Address.String(),
                "mtu":         reply.LinkMtu,
            }
            if reply.Flags&vppintftypes.IF_STATUS_API_FLAG_ADMIN_UP == 0 {
                iface["state"] = "down"
            } else {
                iface["state"] = "up"
            }
            interfaces[name] = iface
        }

        // Output sesuai format contoh Anda (hanya bagian interfaces)
        c.JSON(http.StatusOK, gin.H{
            "interfaces": interfaces,
        })
    }
}
