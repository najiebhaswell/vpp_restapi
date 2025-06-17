package _interface

import (
    "github.com/gin-gonic/gin"
    "vpp-restapi/internal/api"
    vppintf "vpp-restapi/binapi/interface"
    vppintftypes "vpp-restapi/binapi/interface_types"
    lcptype "vpp-restapi/binapi/lcpng_if"
    bondapi "vpp-restapi/binapi/bond"
    "net/http"
    "strings"
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

        // --- Ambil mapping LCP Pair ---
        lcpReq := &lcptype.LcpItfPairGet{}
        lcpReply := &lcptype.LcpItfPairDetails{}
        lcpCtx := ch.SendMultiRequest(lcpReq)
        lcpMap := map[uint32]string{}
        hostIfMap := map[uint32]string{} // PATCH: simpan mapping sw_if_index -> host_if_name
        for {
            stop, err := lcpCtx.ReceiveReply(lcpReply)
            if err != nil || stop {
                break
            }
            lcpMap[uint32(lcpReply.PhySwIfIndex)] = lcpReply.HostIfName
            hostIfMap[uint32(lcpReply.PhySwIfIndex)] = lcpReply.HostIfName // PATCH
            // lcpReply.HostSwIfIndex juga dapat dipakai jika ingin mapping dua arah
        }

        // --- Ambil data bond ---
        bondReq := &bondapi.SwBondInterfaceDump{}
        bondReply := &bondapi.SwBondInterfaceDetails{}
        bondCtx := ch.SendMultiRequest(bondReq)
        bondMap := map[uint32]*bondapi.SwBondInterfaceDetails{}
        for {
            stop, err := bondCtx.ReceiveReply(bondReply)
            if err != nil || stop {
                break
            }
            // Copy struct to avoid pointer overwrite
            copy := *bondReply
            bondMap[uint32(bondReply.SwIfIndex)] = &copy
        }

        // --- Ambil semua member untuk setiap bond ---
        bondMembers := map[uint32][]string{}
        for swif := range bondMap {
            memReq := &bondapi.SwMemberInterfaceDump{SwIfIndex: vppintftypes.InterfaceIndex(swif)}
            memReply := &bondapi.SwMemberInterfaceDetails{}
            memCtx := ch.SendMultiRequest(memReq)
            var members []string
            for {
                stop, err := memCtx.ReceiveReply(memReply)
                if err != nil || stop {
                    break
                }
                members = append(members, memReply.InterfaceName)
            }
            bondMembers[swif] = members
        }

        // --- Ambil interface ---
        req := &vppintf.SwInterfaceDump{}
        reply := &vppintf.SwInterfaceDetails{}
        reqCtx := ch.SendMultiRequest(req)

        interfaces := map[string]map[string]interface{}{}

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

            // --- Tambahkan lcp jika ada ---
            if lcp, ok := lcpMap[uint32(reply.SwIfIndex)]; ok && lcp != "" {
                iface["lcp"] = lcp
                iface["host_if_name"] = lcp // PATCH: tampilkan nama tap/host interface hasil mirror
            }

            // PATCH: Jika interface adalah tap hasil mirror, tambahkan host_if_name juga
            for _, hostIfName := range hostIfMap {
                if name == hostIfName {
                    iface["is_host_if"] = true
                }
            }

            // --- Tambahkan detail bond jika ini bond ---
            if b, ok := bondMap[uint32(reply.SwIfIndex)]; ok {
                iface["bond_id"] = b.ID
                iface["bond_mode"] = b.Mode.String()
                iface["bond_load_balance"] = b.Lb.String()
                iface["bond_members"] = bondMembers[uint32(reply.SwIfIndex)]
            }

            // --- Subinterface logic (seperti sebelumnya) ---
            if idx := strings.Index(name, "."); idx > 0 {
                parent := name[:idx]
                subid := name[idx+1:]

                ifaceParent, ok := interfaces[parent]
                if !ok {
                    ifaceParent = map[string]interface{}{
                        "description":    "",
                        "mac":            "",
                        "mtu":            0,
                        "state":          "",
                        "sub-interfaces": map[string]interface{}{},
                    }
                    interfaces[parent] = ifaceParent
                }
                subif, ok := ifaceParent["sub-interfaces"].(map[string]interface{})
                if !ok || subif == nil {
                    subif = map[string]interface{}{}
                    ifaceParent["sub-interfaces"] = subif
                }
                subif[subid] = iface
            } else {
                interfaces[name] = iface
            }
        }

        c.IndentedJSON(http.StatusOK, gin.H{
            "interfaces": interfaces,
        })
    }
}
