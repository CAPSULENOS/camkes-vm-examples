package build

import (
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "orchestration/helpers"
    "orchestration/types"
)

type FunctionalityInfo = types.FunctionalityInfo
type NodeInfo = types.NodeInfo
type DebugInfo  = types.DebugInfo
type Config  = types.Config


func EtcHosts(available_nodes []NodeInfo) error {

    // Open /etc/hosts for ease of use
    hosts, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        helpers.LogE("Error opening /etc/hosts:", err)
        return err
    }
    defer hosts.Close()

    // Actually write sel4 names to etc hosts
    for _, node := range available_nodes {
        hosts.Write([]byte(node.Sel4IP + " " + node.Sel4Name + "\n"))
    }
    return nil
}

func FunctionsOnNode(name string, settings FunctionalityInfo) error {

    switch settings.Type {
        case "router":
            settings.BuildRouterNode()
        case "silent":
            settings.BuildSilentNode()
        default:
            helpers.LogE("Invalid vnf type:", settings.Type)
            return nil
    }

    return nil
}

func ForwardingPath(connections [][2]int, Type string) error {

    // Apply ingest path rules 
    fmt.Println("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", connections[0][0], connections[0][1]))
    cmd := exec.Command("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", connections[0][0], connections[0][1]))
    output, err := cmd.CombinedOutput()
    if err != nil {
        helpers.LogE("Error creating ingest arp : ", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", connections[0][0], connections[0][1]), string(output), "Exiting")
        return err
    }
    fmt.Println("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("in_port=eth1,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth0", connections[0][0], connections[0][1]))
    cmd = exec.Command("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("in_port=eth1,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth0", connections[0][0], connections[0][1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        helpers.LogE("Error creating ingest arp : ", fmt.Sprintf("in_port=eth1,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth0", connections[0][0], connections[0][1]), string(output), "Exiting")
        return err
    }

    internal_connections := connections[1:]
    if Type == "bidirectional" { 
        internal_connections = internal_connections[:len(connections)-2]; 
    }

    // Apply the forwarding paths on the current node
    for _, connection := range internal_connections {
        fmt.Println("/connect-vlans", "--uni", "--vid1", strconv.Itoa(connection[0]), "--vid2", strconv.Itoa(connection[1]))
        cmd = exec.Command("/root/connect-vlans", "--uni", "--vid1", strconv.Itoa(connection[0]), "--vid2", strconv.Itoa(connection[1]))
        output, err = cmd.CombinedOutput()
        if err != nil {
            helpers.LogE("Error connecting vlans: ", "/root/connect-vlans", "--uni", "--vid1", strconv.Itoa(connection[0]), "--vid2", strconv.Itoa(connection[1]), string(output), "Exiting")
            return err
        }

    }

    if Type == "unidirectional" { return nil }

    // Apply egress rules
    egress_connection := connections[len(connections)-1]
    // Apply ingest path rules 
    fmt.Println("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", egress_connection[0], egress_connection[1]))
    cmd = exec.Command("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", egress_connection[0], egress_connection[1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        helpers.LogE("Error creating ingest arp : ", fmt.Sprintf("arp,dl_vlan=%d,action=mod_vlan_vid=%d,NORMAL", egress_connection[0], egress_connection[1]), string(output), "Exiting")
        return err
    }
    fmt.Println("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("in_port=eth0,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth1", egress_connection[0], egress_connection[1]))
    cmd = exec.Command("ovs-ofctl", "add-flow", "br0", fmt.Sprintf("in_port=eth0,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth1", egress_connection[0], egress_connection[1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        helpers.LogE("Error creating ingest arp : ", fmt.Sprintf("in_port=eth0,dl_vlan=%v,actions=mod_vlan_vid=%v,output=eth1", egress_connection[0], egress_connection[1]), string(output), "Exiting")
        return err
    }


    return nil
}

