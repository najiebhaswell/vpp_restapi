package _interface

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    vppinterface "vpp-restapi/binapi/interface"
    vppinterface_types "vpp-restapi/binapi/interface_types"
    vppbond "vpp-restapi/binapi/bond"
    vppip "vpp-restapi/binapi/ip"
    vppapi "vpp-restapi/internal/api"
    lcptype "vpp-restapi/binapi/lcpng_if"
)

// RegisterRoutes sets up the interface-related HTTP routes.
func RegisterRoutes(r gin.IRoutes, vppClient *vppapi.VPPClient) {
    r.GET("/interfaces", listInterfacesHandler(vppClient))
    r.POST("/interfaces/loopback", createLoopbackHandler(vppClient))
    r.DELETE("/interfaces/:sw_if_index", deleteInterfaceHandler(vppClient))
    r.POST("/interfaces/:sw_if_index/enable", enableInterfaceHandler(vppClient))
    r.POST("/interfaces/:sw_if_index/disable", disableInterfaceHandler(vppClient))
}

// Handler untuk membuat loopback interface
func createLoopbackHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        var config struct {
            MacAddress string `json:"mac_address"`
        }
        if err := c.ShouldBindJSON(&config); err != nil {
            log.Printf("Invalid request body: %v", err)
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

        var mac [6]byte
        if config.MacAddress != "" {
            if err := parseMacAddress(config.MacAddress, &mac); err != nil {
                log.Printf("Invalid MAC address: %v", err)
                c.JSON(http.StatusBadRequest, gin.H{"error": "invalid MAC address", "details": err.Error()})
                return
            }
        }

        req := &vppinterface.CreateLoopback{
            MacAddress: mac,
        }
        reply := &vppinterface.CreateLoopbackReply{}
        if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
            log.Printf("API request failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed", "details": err.Error()})
            return
        }
        if reply.Retval != 0 {
            log.Printf("Loopback creation failed with retval: %d", reply.Retval)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "loopback creation failed", "details": fmt.Sprintf("VPP returned non-zero retval: %d", reply.Retval)})
            return
        }

        log.Printf("Created loopback with SwIfIndex=%d", reply.SwIfIndex)
        c.JSON(http.StatusCreated, gin.H{"message": "Loopback created", "sw_if_index": reply.SwIfIndex})
    }
}

// Handler untuk menghapus interface
func deleteInterfaceHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        swIfIndex, err := strconv.Atoi(c.Param("sw_if_index"))
        if err != nil {
            log.Printf("Invalid sw_if_index: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sw_if_index", "details": err.Error()})
            return
        }

        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            log.Printf("Failed to create API channel: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "channel creation failed", "details": err.Error()})
            return
        }
        defer ch.Close()

        // Check interface type
        reqDump := &vppinterface.SwInterfaceDump{SwIfIndex: vppinterface_types.InterfaceIndex(swIfIndex)}
        replyDump := &vppinterface.SwInterfaceDetails{}
        reqCtx := ch.SendMultiRequest(reqDump)
        isBond := false
        for {
            stop, err := reqCtx.ReceiveReply(replyDump)
            if err != nil {
                log.Printf("Dump request failed: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "dump request failed", "details": err.Error()})
                return
            }
            if stop {
                break
            }
            if strings.HasPrefix(string(replyDump.InterfaceName[:]), "BondEthernet") {
                isBond = true
            }
        }

        if isBond {
            req := &vppbond.BondDelete{
                SwIfIndex: vppinterface_types.InterfaceIndex(swIfIndex),
            }
            reply := &vppbond.BondDeleteReply{}
            if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
                log.Printf("API request failed: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed", "details": err.Error()})
                return
            }
            if reply.Retval != 0 {
                log.Printf("Bond deletion failed with retval: %d", reply.Retval)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "bond deletion failed", "details": fmt.Sprintf("VPP returned non-zero retval: %d", reply.Retval)})
                return
            }
            log.Printf("Deleted bond interface %d", swIfIndex)
            c.JSON(http.StatusOK, gin.H{"message": "Bond deleted"})
        } else {
            req := &vppinterface.DeleteLoopback{
                SwIfIndex: vppinterface_types.InterfaceIndex(swIfIndex),
            }
            reply := &vppinterface.DeleteLoopbackReply{}
            if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
                log.Printf("API request failed: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed", "details": err.Error()})
                return
            }
            if reply.Retval != 0 {
                log.Printf("Loopback deletion failed with retval: %d", reply.Retval)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "loopback deletion failed", "details": fmt.Sprintf("VPP returned non-zero retval: %d", reply.Retval)})
                return
            }
            log.Printf("Deleted loopback interface %d", swIfIndex)
            c.JSON(http.StatusOK, gin.H{"message": "Loopback deleted"})
        }
    }
}

// Handler untuk enable interface
func enableInterfaceHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        swIfIndex, err := strconv.Atoi(c.Param("sw_if_index"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sw_if_index", "details": err.Error()})
            return
        }
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        req := &vppinterface.SwInterfaceSetFlags{
            SwIfIndex: vppinterface_types.InterfaceIndex(swIfIndex),
            Flags:     vppinterface_types.IF_STATUS_API_FLAG_ADMIN_UP,
        }
        reply := &vppinterface.SwInterfaceSetFlagsReply{}
        if err := ch.SendRequest(req).ReceiveReply(reply); err != nil || reply.Retval != 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable interface", "details": err})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Interface enabled"})
    }
}

// Handler untuk disable interface
func disableInterfaceHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        swIfIndex, err := strconv.Atoi(c.Param("sw_if_index"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sw_if_index", "details": err.Error()})
            return
        }
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer ch.Close()

        req := &vppinterface.SwInterfaceSetFlags{
            SwIfIndex: vppinterface_types.InterfaceIndex(swIfIndex),
            Flags:     0, // Admin down
        }
        reply := &vppinterface.SwInterfaceSetFlagsReply{}
        if err := ch.SendRequest(req).ReceiveReply(reply); err != nil || reply.Retval != 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable interface", "details": err})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Interface disabled"})
    }
}

// Helper to convert array IP to net.IP
func ip4ToNetIP(a [4]byte) net.IP {
    return net.IP{a[0], a[1], a[2], a[3]}
}
func ip6ToNetIP(a [16]byte) net.IP {
    return net.IP{
        a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7],
        a[8], a[9], a[10], a[11], a[12], a[13], a[14], a[15],
    }
}

// @Summary List all interfaces
// @Description Get all VPP interfaces with status, IP addresses, bond config, and host_if_name if mirrored via LCP.
// @Tags interfaces
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /vpp/interfaces [get]
func listInterfacesHandler(vppClient *vppapi.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            log.Printf("Failed to create API channel: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "channel creation failed", "details": err.Error()})
            return
        }
        defer ch.Close()

        // --- Ambil mapping LCP (phySwIfIndex -> hostIfName) ---
        lcpReq := &lcptype.LcpItfPairGet{}
        lcpReply := &lcptype.LcpItfPairDetails{}
        lcpCtx := ch.SendMultiRequest(lcpReq)
        lcpMap := map[uint32]string{}
        for {
            stop, err := lcpCtx.ReceiveReply(lcpReply)
            if err != nil || stop {
                break
            }
            lcpMap[uint32(lcpReply.PhySwIfIndex)] = lcpReply.HostIfName
        }

        // --- Ambil data bond ---
        bondReq := &vppbond.SwBondInterfaceDump{}
        bondReply := &vppbond.SwBondInterfaceDetails{}
        bondCtx := ch.SendMultiRequest(bondReq)
        bondMap := map[uint32]*vppbond.SwBondInterfaceDetails{}
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
            memReq := &vppbond.SwMemberInterfaceDump{SwIfIndex: vppinterface_types.InterfaceIndex(swif)}
            memReply := &vppbond.SwMemberInterfaceDetails{}
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

        interfaces := []map[string]interface{}{}
        req := &vppinterface.SwInterfaceDump{}
        reply := &vppinterface.SwInterfaceDetails{}
        reqCtx := ch.SendMultiRequest(req)
        for {
            stop, err := reqCtx.ReceiveReply(reply)
            if err != nil {
                log.Printf("API request failed: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed", "details": err.Error()})
                return
            }
            if stop {
                break
            }
            interfaceType := inferInterfaceType(reply)

            ipAddresses := []string{}
            for _, isIPv6 := range []bool{false, true} {
                ipCh, err := vppClient.NewAPIChannel()
                if err != nil {
                    continue
                }
                defer ipCh.Close()

                ipReq := &vppip.IPAddressDump{
                    SwIfIndex: reply.SwIfIndex,
                    IsIPv6:    isIPv6,
                }
                ipReply := &vppip.IPAddressDetails{}
                ipCtx := ipCh.SendMultiRequest(ipReq)
                for {
                    stopAddr, err := ipCtx.ReceiveReply(ipReply)
                    if err != nil {
                        break // skip jika error
                    }
                    if stopAddr {
                        break
                    }
                    prefix := ipReply.Prefix
                    var ipnet net.IPNet
                    if isIPv6 {
                        ipnet.IP = ip6ToNetIP(prefix.Address.Un.GetIP6())
                        ipnet.Mask = net.CIDRMask(int(prefix.Len), 128)
                    } else {
                        ipnet.IP = ip4ToNetIP(prefix.Address.Un.GetIP4())
                        ipnet.Mask = net.CIDRMask(int(prefix.Len), 32)
                    }
                    ipAddresses = append(ipAddresses, ipnet.String())
                }
            }

            iface := map[string]interface{}{
                "index":        uint32(reply.SwIfIndex),
                "name":         string(reply.InterfaceName[:]),
                "type":         interfaceType,
                "admin_up":     (reply.Flags & vppinterface_types.IF_STATUS_API_FLAG_ADMIN_UP) != 0,
                "link_up":      (reply.Flags & vppinterface_types.IF_STATUS_API_FLAG_LINK_UP) != 0,
                "mtu":          reply.Mtu[0], // Use L3 MTU
                "ip_addresses": ipAddresses,
            }

            // PATCH: Tambahkan host_if_name jika ada mapping LCP
            if lcp, ok := lcpMap[uint32(reply.SwIfIndex)]; ok && lcp != "" {
                iface["host_if_name"] = lcp
            }

            // PATCH: Tambahkan konfigurasi bond jika interface ini bond
            if b, ok := bondMap[uint32(reply.SwIfIndex)]; ok {
                iface["bond_id"] = b.ID
                iface["bond_mode"] = b.Mode.String()
                iface["bond_load_balance"] = b.Lb.String()
                iface["bond_members"] = bondMembers[uint32(reply.SwIfIndex)]
            }
            // END PATCH

            interfaces = append(interfaces, iface)
        }

        c.JSON(http.StatusOK, gin.H{"interfaces": interfaces})
    }
}

// inferInterfaceType infers the interface type based on SwInterfaceDetails.
func inferInterfaceType(details *vppinterface.SwInterfaceDetails) string {
    name := string(details.InterfaceName[:])
    switch {
    case name == "local0":
        return "local"
    case details.LinkDuplex != 0:
        return "ethernet"
    default:
        return "unknown"
    }
}

// parseMacAddress parses a MAC address string into a [6]byte array.
func parseMacAddress(macStr string, mac *[6]byte) error {
    parts := strings.Split(macStr, ":")
    if len(parts) != 6 {
        return fmt.Errorf("invalid MAC address format: %s", macStr)
    }
    for i, part := range parts {
        val, err := strconv.ParseUint(part, 16, 8)
        if err != nil {
            return fmt.Errorf("invalid MAC address part %s: %v", part, err)
        }
        mac[i] = byte(val)
    }
    return nil
}
