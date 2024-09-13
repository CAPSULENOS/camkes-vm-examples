package types;

import (
    "os/exec"
    "fmt"
    "strconv"
    // "strings"
)


type Config struct {
    NumberOfVms                           int                                     `yaml:"local_vms"`
    Functionality                         map[string]map[string]FunctionalityInfo `yaml:"functions"`
    DataPaths                             map[string]DataPath                     `yaml:"data_paths"`
    Ingest                                map[string]IngestInfo                   `yaml:"ingest"`
    Debug                                 DebugInfo                               `yaml:"debug"`
}

type FunctionalityInfo struct {
    Name            string
    Type            string // Wireguard, router, silent, etc
    // Routes: to, via
    // Interfaces: dev, addr
    Routes          []ForwardingPath               `yaml:"routes"`
    Interfaces      []Interface                    `yaml:"interfaces"`
    Vlans           [][2]int
    // For assigning to physical nodes
    PhysicalNode   *NodeInfo
}

type DataPath struct {
    Ingest          IngestInfo                     `yaml:"ingest"`
    Type            string                         `yaml:"type"`
    Path            []string                       `yaml:"path"`
    Connections     [][2]int
}

type DebugInfo struct {
    OrderVlans      bool                           `yaml:"order_vlans"`
    Verbose         bool                           `yaml:"verbose"`
}

type NodeInfo struct {
    Name     string
    Mac      string
    Sel4Name string
    Sel4IP   string
}

type ForwardingPath struct {
    To       string                                 `yaml:"to"`
    Via      string                                 `yaml:"via"`
}

type Interface struct {
    Dev     string                                  `yaml:"dev"`
    Addr    string                                  `yaml:"addr"`
    Vlan    int
}

type IngestInfo struct {
    Vlan     int                                    `yaml:"vlan"`
    Subnet   string                                 `yaml:"subnet"`
}


// Below is how all the VNFs actually get implemented


type BuildRouterError struct {
    s string
}

func (e *BuildRouterError) Error() string {
    return e.s
}

func NewBuildRouterError(message string) *BuildRouterError {
    return &BuildRouterError{s: message}
}

func (f FunctionalityInfo) BuildRouterNode() error {

    /*
    // Build a map of subnet to IP address
    subnetToIp := make(map[string]string)
    for _, link := range f.ForwardingPaths {
        subnetToIp[link.From] = link.IP
    }

    */

    node := f.PhysicalNode.Sel4IP

    // Make interfaces using vlan to subnet mapping
    for _, ethX := range f.Interfaces {

        if ethX.Vlan == 0 {
            return fmt.Errorf("Interface for router %v is equal to zero. This is not allowed. Is everything defined correctly?", f.Name)
        }


        ip := ethX.Addr
        if ip == "" { 
            return NewBuildRouterError(fmt.Sprintf("IP set in router not in subnet assigned to ingest path in subnet %v", ip)); 
        }

        fmt.Println("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", ip, "--vid", strconv.Itoa(ethX.Vlan))
        cmd := exec.Command("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", ip, "--vid", strconv.Itoa(ethX.Vlan))
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
            return err
        }
    }

    // Set up arp forwarding on the node
    cmd_text := fmt.Sprintf("/etc/aegis/scripts/arp-forwarding --dev eth0")
    fmt.Println("sshpass", "-p", "root", "ssh", node, cmd_text)

    cmd := exec.Command("sshpass", "-p", "root", "ssh", node, cmd_text)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding arp forwarding on remote node: ", string(output), "Exiting")
        return err
    }

    // Set up the forwarding path on the remote host
    for _, link := range f.Routes {
        fmt.Println("/etc/aegis/scripts/add-route", "--node", node, "--to", link.To, "--via", link.Via)
        cmd := exec.Command("/etc/aegis/scripts/add-route", "--node", node, "--to", link.To, "--via", link.Via)
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error adding route on remote node: ", string(output), err, "Exiting")
            return err
        }
    }


    return nil
}

/*
type BuildSilentError struct {
    s string
}

func (e *BuildSilentError) Error() string {
    return e.s
}

func NewBuildSilentError(message string) *BuildRouterError {
    return &BuildRouterError{s: message}
}
*/

func (f FunctionalityInfo) BuildSilentNode() error {
    node := f.PhysicalNode.Sel4IP

    for _, vlans := range f.Vlans {
        fmt.Println("/etc/aegis/scripts/silent", "--uni", "--node", node, "--dev", "eth0", "--vid1", strconv.Itoa(vlans[0]), "--vid2", strconv.Itoa(vlans[1]))
        cmd := exec.Command("/etc/aegis/scripts/silent", "--uni", "--node", node, "--dev", "eth0", "--vid1", strconv.Itoa(vlans[0]), "--vid2", strconv.Itoa(vlans[1]))
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error implementing a silent node: ", string(output), "Exiting")
            return err
        }
    }
    
    return nil
}

