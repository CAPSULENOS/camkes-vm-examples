package types;

import (
    "os/exec"
    "fmt"
    "strconv"
    "strings"
)


type Config struct {
    Functionality                         map[string]map[string]FunctionalityInfo `yaml:"functions"`
    DataPaths                             map[string]DataPath            `yaml:"data_paths"`
    Debug                                 DebugInfo                      `yaml:"debug"`
}

type FunctionalityInfo struct {
    Name            string    
    Type            string // Wireguard, router, silent, etc
    SubnetToVlan    map[string]int
    ForwardingPaths []ForwardingPath               `yaml:"links"`
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
    From     string                                 `yaml:"from"`
    IP       string                                 `yaml:"ip"`
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

    // Build a map of subnet to IP address
    subnetToIp := make(map[string]string)
    for _, link := range f.ForwardingPaths {
        subnetToIp[link.From] = link.IP
    }

    node := f.PhysicalNode.Sel4IP

    // Make interfaces using vlan to subnet mapping
    for subnet, vlan := range f.SubnetToVlan {
        ip := subnetToIp[subnet]

        if ip == "" { 
            return NewBuildRouterError(fmt.Sprintf("IP set in router not in subnet assigned to ingest path in subnet %v", subnet)); 
        }

        fmt.Println("/root/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", ip, "--vid", strconv.Itoa(vlan))
        cmd := exec.Command("/root/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", ip, "--vid", strconv.Itoa(vlan))
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
            return err
        }
     }

    // Set up the forwarding path on the remote host
    for _, link := range f.ForwardingPaths {
        raw_ip := strings.Split(link.IP, "/")[0]
        fmt.Println("/root/add-route", "--node", node, "--to", link.To, "--via", raw_ip)
        cmd := exec.Command("/root/add-route", "--node", node, "--to", link.To, "--via", raw_ip)
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error adding route on remote node: ", string(output), "Exiting")
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
        fmt.Println("/root/silent", "--uni", "--node", node, "--dev", "eth0", "--vid1", strconv.Itoa(vlans[0]), "--vid2", strconv.Itoa(vlans[1]))
        cmd := exec.Command("/root/silent", "--uni", "--node", node, "--dev", "eth0", "--vid1", strconv.Itoa(vlans[0]), "--vid2", strconv.Itoa(vlans[1]))
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error implementing a silent node: ", string(output), "Exiting")
            return err
        }
    }
    
    return nil
}

