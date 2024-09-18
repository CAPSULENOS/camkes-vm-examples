package takedown;

import (
    "os/exec"
    "fmt"
    "orchestration/nodes"
)


func StopAegis(lease_file string) error {
    // Get all the devices
    available_nodes, err := nodes.GetAvailableNodes(lease_file)


    // Make all the devices clean themselves
    for _, node := range available_nodes {
        cmd_text := "/etc/aegis/scripts/clean"

        fmt.Println("sshpass", "-p", "root", "dbclient", "-y", node.Sel4IP, cmd_text)
        cmd := exec.Command("sshpass", "-p", "root", "dbclient", "-y", node.Sel4IP, cmd_text)
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error adding arp forwarding on remote node: ", string(output), "Exiting")
            return err
        }
    }


    // Remove all local ovs rules
    fmt.Println("ovs-ofctl", "del-flows", "br0")
    cmd := exec.Command("ovs-ofctl", "del-flows", "br0")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error removing openvswitch flows: ", err, string(output), "Exiting")
        return fmt.Errorf("Error removing openvswitch flows: ", err, string(output), "Exiting")
    }
 

    // Add back in basic forwarding rules
    fmt.Println("ovs-ofctl", "add-flow", "br0", "priority=100,in_port=eth0 actions=output:br0")
    cmd = exec.Command("ovs-ofctl", "add-flow", "br0", "priority=100,in_port=eth0 actions=output:br0")
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding openvswitch flows: ", err, string(output), "Exiting")
        return fmt.Errorf("Error adding openvswitch flows: ", err, string(output), "Exiting")
    }
 
    fmt.Println("ovs-ofctl", "add-flow", "br0", "priority=0,actions=output:eth0")
    cmd = exec.Command("ovs-ofctl", "add-flow", "br0", "priority=0,actions=output:eth0")
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error addding openvswitch flows: ", err, string(output), "Exiting")
        return fmt.Errorf("Error adding openvswitch flows: ", err, string(output), "Exiting")
    }


    return nil
}

