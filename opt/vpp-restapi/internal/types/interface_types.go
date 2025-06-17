package types

import "fmt"

func InferInterfaceType(swIfIndex uint32) (string, string) {
    if swIfIndex == 0 {
        return "local", "local0"
    }
    switch swIfIndex {
    case 1:
        return "loopback", "loop0"
    case 2:
        return "bond", "BondEthernet0"
    case 3:
        return "tap", "tap0"
    case 4:
        return "loopback", "loop1"
    default:
        return "unknown", fmt.Sprintf("interface-%d", swIfIndex)
    }
}
