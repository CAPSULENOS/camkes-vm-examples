package types;

import (
    "os/exec"
    "fmt"
    "strconv"
    "strings"
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
    // For router
    Routes          []ForwardingPath               `yaml:"routes"`
    Interfaces      []Interface                    `yaml:"interfaces"`
    // For silent
    Vlans           [][2]int
    // For wireguard
    IngestAddr      string                         `yaml:"ingest_ip"`
    EgressAddr      string                         `yaml:"egress_ip"`
    WgInterface     WgInterface                    `yaml:"interface"`
    Peers           map[string]WgPeer              `yaml:"peers"`
    // For iptables
    Setup           []string                       `yaml:"setup"`
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
    Vlan    int                                     `yaml:"vlan"`
}

type IngestInfo struct {
    Vlan     int                                    `yaml:"vlan"`
    Subnet   string                                 `yaml:"subnet"`
}

type WgInterface struct {
    Name         string                             `yaml:"name"`
    PrivateKey   string                             `yaml:"private_key"`
    Addr         string                             `yaml:"address"`
    Port         string                             `yaml:"port"`
}

type WgPeer struct {
    PublicKey    string                             `yaml:"public_key"`
    AllowedIPs   []string                           `yaml:"allowed_ips"`
    Endpoint     string                             `yaml:"endpoint"`
}

// Below is how all the VNFs actually get implemented


func (f FunctionalityInfo) BuildRouterNode() error {
    node := f.PhysicalNode.Sel4IP

    // Make interfaces using vlan to subnet mapping
    for _, ethX := range f.Interfaces {

        if ethX.Vlan == 0 {
            return fmt.Errorf("Interface for router %v is equal to zero. This is not allowed. Is everything defined correctly?", f.Name)
        }


        ip := ethX.Addr
        if ip == "" { 
            return fmt.Errorf("IP set in router not in subnet assigned to ingest path in subnet %v", ip); 
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

    // Add ipv4 forwarding on a remote host
    fmt.Println("/etc/aegis/scripts/add-ip-forward", "--node", node)
    cmd = exec.Command("/etc/aegis/scripts/add-ip-forward", "--node", node)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding ipv4 forwarding on remote node: ", string(output), err, "Exiting")
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


func GetWireGuardInterfaceName(settings FunctionalityInfo, path_num int) string {
    if settings.WgInterface.Name == "" { settings.WgInterface.Name = "wg" + strconv.Itoa(path_num); }
    return settings.WgInterface.Name
}


var path_num int;

func (f FunctionalityInfo) BuildWireGuardNode() error {
    node := f.PhysicalNode.Sel4IP


    // Set up interface to get packets from
    fmt.Println("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", f.IngestAddr, "--vid", strconv.Itoa(f.Vlans[0][0]))
    cmd := exec.Command("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", f.IngestAddr, "--vid", strconv.Itoa(f.Vlans[0][0]))
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
        return err
    }


    // Set up interface to forward packets from (in real network)
    fmt.Println("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", f.EgressAddr, "--vid", strconv.Itoa(f.Vlans[0][1]))
    cmd = exec.Command("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--ip", f.EgressAddr, "--vid", strconv.Itoa(f.Vlans[0][1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
        return err
    }


    // Set the egress interface as the default interface
    ip := strings.Split(f.EgressAddr,"/")[0]
    fmt.Println("/etc/aegis/scripts/set-default-route", "--node", node, "--dev", fmt.Sprintf("eth0.%v", strconv.Itoa(f.Vlans[0][1])), "--addr", ip)
    cmd = exec.Command("/etc/aegis/scripts/set-default-route", "--node", node, "--dev", fmt.Sprintf("eth0.%v", strconv.Itoa(f.Vlans[0][1])), "--addr", ip)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error setting default route on client: ", string(output), "Exiting")
        return err
    }
 

    // Enable IPv4 forwarding
    fmt.Println("/etc/aegis/scripts/add-ip-forward", "--node", node)
    cmd = exec.Command("/etc/aegis/scripts/add-ip-forward", "--node", node)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding ipv4 forwarding on remote node: ", string(output), err, "Exiting")
        return err
    }


    // Set up arp forwarding on the node
    cmd_text := fmt.Sprintf("/etc/aegis/scripts/arp-forwarding --dev eth0")
    fmt.Println("sshpass", "-p", "root", "ssh", node, cmd_text)

    cmd = exec.Command("sshpass", "-p", "root", "ssh", node, cmd_text)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding arp forwarding on remote node: ", string(output), "Exiting")
        return err
    }



    interface_name := GetWireGuardInterfaceName(f, path_num)
    path_num += 1

    fmt.Println("Setting up wireguard interface ", interface_name)

    fmt.Println("/etc/aegis/scripts/wireguard-interface.sh", "--interface", interface_name, "--node", node,
                            "--pk", f.WgInterface.PrivateKey, "--addr", f.WgInterface.Addr, "--port", f.WgInterface.Port)
    cmd = exec.Command("/etc/aegis/scripts/wireguard-interface.sh", "--interface", interface_name,  "--node", node,
                                "--pk", f.WgInterface.PrivateKey, "--addr", f.WgInterface.Addr, "--port", f.WgInterface.Port)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error creating wireguard interface: ", interface_name, err, string(output), "Exiting")
        return fmt.Errorf("Error creating wireguard interface: %v\n%v\n%v\nExiting", interface_name, err, output)
    }
 

    // Add in peers so we can do forwarding
    for name, peer_settings := range f.Peers {
        // Build allowed ips argument string
        allowed_ips := ""
        for _, ip := range peer_settings.AllowedIPs { 
            // ip = ip[:len(ip) - 3]
            allowed_ips += ip + "," 
        }
        // remove trailing comma
        allowed_ips = allowed_ips[:len(allowed_ips) - 1]
        fmt.Println("sshpass", "-p", "root", "dbclient", "-y", node,
            "wg", "set", interface_name, "peer", peer_settings.PublicKey, 
                                "allowed-ips", allowed_ips, "endpoint", peer_settings.Endpoint)

        cmd := exec.Command("sshpass", "-p", "root", "dbclient", "-y", node,
            "wg", "set", interface_name, "peer", peer_settings.PublicKey, 
                                "allowed-ips", allowed_ips, "endpoint", peer_settings.Endpoint)

        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Error creating wireguard peer: ", name, err, string(output), "Exiting")
            return fmt.Errorf("Error creating wireguard peer: %v\n%v\n%v\nExiting", name, err, output)
        }

    }


    // Add in routes so we know how to forward via wireguard
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

func (f FunctionalityInfo) BuildIptablesNode() error {
    node := f.PhysicalNode.Sel4IP

    // Set up ingest interface
    fmt.Println("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--vid", strconv.Itoa(f.Vlans[0][0]))
    cmd := exec.Command("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--vid", strconv.Itoa(f.Vlans[0][0]))
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
        return err
    }

    // Set up egress interface
    fmt.Println("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--vid", strconv.Itoa(f.Vlans[0][1]))
    cmd = exec.Command("/etc/aegis/scripts/generate-client-vlan", "--node", node, "--dev", "eth0", "--vid", strconv.Itoa(f.Vlans[0][1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error generating vlans on client: ", string(output), "Exiting")
        return err
    }

    // Set up arp forwarding on the node
    cmd_text := fmt.Sprintf("/etc/aegis/scripts/arp-forwarding --dev eth0")
    fmt.Println("sshpass", "-p", "root", "ssh", node, cmd_text)

    cmd = exec.Command("sshpass", "-p", "root", "ssh", node, cmd_text)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding arp forwarding on remote node: ", string(output), "Exiting")
        return err
    }

    // Enable IPv4 forwarding
    fmt.Println("/etc/aegis/scripts/add-ip-forward", "--node", node)
    cmd = exec.Command("/etc/aegis/scripts/add-ip-forward", "--node", node)
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding ipv4 forwarding on remote node: ", string(output), err, "Exiting")
        return err
    }

    // Enable forwarding between the VLANs using iptables
    fmt.Println("/etc/aegis/scripts/iptables-raw-forwarding", "--node", node, "--vid1", strconv.Itoa(f.Vlans[0][0]), "--vid2", strconv.Itoa(f.Vlans[0][1]))
    cmd = exec.Command("/etc/aegis/scripts/iptables-raw-forwarding", "--node", node, "--vid1", strconv.Itoa(f.Vlans[0][0]), "--vid2", strconv.Itoa(f.Vlans[0][1]))
    output, err = cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error adding iptables raw forwarding on remote node: ", string(output), err, "Exiting")
        return err
    }




    return nil
}

