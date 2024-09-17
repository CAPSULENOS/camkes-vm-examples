package wireguard

import (
    "os"
    "fmt"
    "os/exec"
    "strconv"
    "orchestration/types"
    "orchestration/helpers"
)

type DebugInfo = types.DebugInfo
type IngestInfo = types.IngestInfo


func GetWireGuardInterfaceName(settings IngestInfo, path_num int) string {
    if settings.InterfaceName == "" { settings.InterfaceName = "wg" + strconv.Itoa(path_num); }
    return settings.InterfaceName
}

func StartWireGuard(settings IngestInfo, path_num int, debug DebugInfo) bool {

    interface_name := GetWireGuardInterfaceName(settings, path_num)

    if debug.Verbose { helpers.LogE("Setting up wireguard interface ", interface_name) }

    if debug.Verbose { 
        fmt.Println("/root/wireguard-interface.sh", "--interface", interface_name, 
                                "--pk", settings.PrivateKey, "--addr", settings.Address, "--port", settings.ListenPort)
    }
    cmd := exec.Command("/root/wireguard-interface.sh", "--interface", interface_name, 
                            "--pk", settings.PrivateKey, "--addr", settings.Address, "--port", settings.ListenPort)
    output, err := cmd.CombinedOutput()
    if err != nil {
        helpers.LogE("Error creating wireguard interface: ", interface_name, err, output, "Exiting")
        return true
    }


    for name, peer_settings := range settings.Peers {

        // Build allowed ips argument string
        allowed_ips := ""
        for _, ip := range peer_settings.AllowedIPs { 
            // ip = ip[:len(ip) - 3]
            allowed_ips += ip + "," 
        }
        // remove trailing comma
        allowed_ips = allowed_ips[:len(allowed_ips) - 1]
        if debug.Verbose { 
            fmt.Println("Peer name:", name);
            fmt.Println("wg", "set", interface_name, "peer", peer_settings.PublicKey, 
                                    "allowed-ips", allowed_ips, "endpoint", peer_settings.Endpoint)
        }

        cmd := exec.Command("wg", "set", interface_name, "peer", peer_settings.PublicKey, 
                                "allowed-ips", allowed_ips, "endpoint", peer_settings.Endpoint) 

        output, err := cmd.CombinedOutput()
        if err != nil {
            helpers.LogE("Error creating wireguard peer: ", name, err, output, "Exiting")
            return true
        }

    }

    return false
}


