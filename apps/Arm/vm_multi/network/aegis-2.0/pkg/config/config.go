package config;

import (
    "os"
    "fmt"
    "net"
    "strconv"
    "gopkg.in/yaml.v3"
)


type FunctionType string

const (
    Router FunctionType = "router"
    WireGuard FunctionType = "wireguard"
    IpTables FunctionType = "iptables"
    Silent FunctionType = "silent"
)

func (t FunctionType) isValid() bool {
    switch t {
        case Router, WireGuard, Silent:
            return true
    }
    return false
}


type Directionality string

const (
    Unidirectional Directionality = "unidirectional"
    Bidirectional Directionality = "bidirectional"
)

func (t Directionality) isValid() bool {
    switch t {
        case Unidirectional, Bidirectional:
            return true
    }
    return false
}


type Config struct {
    NumberOfVms                           int                                     `yaml:"local_vms"`
    Functionality                         map[FunctionType]*NamedFunctionality    `yaml:"functions"`
    DataPaths                             map[string]*DataPath                     `yaml:"data_paths"`
    Ingest                                map[string]*IngestInfo                   `yaml:"ingest"`
    Debug                                 *DebugInfo                               `yaml:"debug"`
}

type NamedFunctionality map[string]*FunctionalityInfo;

type FunctionalityInfo struct {
    Name            string
    Type            FunctionType // Wireguard, router, silent, etc
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
    Type            Directionality                 `yaml:"type"`
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
}

type IngestInfo struct {
    Vlan     int                                    `yaml:"vlan"`
}

type WgInterface struct {
    Name         string                             `yaml:"name"`
    PrivateKey   string                             `yaml:"private_key"`
    Addr         string                             `yaml:"address"`
    Port         int                                `yaml:"port"`
}

type WgPeer struct {
    PublicKey    string                             `yaml:"public_key"`
    AllowedIPs   []string                           `yaml:"allowed_ips"`
    Endpoint     string                             `yaml:"endpoint"`
}


func (c *Config) LoadFromFile(filename string) error {
    bytes, err := os.ReadFile(filename)
    if err != nil {
        return err
    }

    err = yaml.Unmarshal(bytes, c)
    if err != nil {
        return err
    }

    return c.check()
}


func (c *Config) check() error {
    var err error;

    
    if c.NumberOfVms == 0 {
        return fmt.Errorf("aegis config 'local_vms' cannot be zero. Please specify the number of VMs running in the node")
    }

    if len(c.Functionality) == 0 {
        return fmt.Errorf("you must declare named functionality in an aegis node")
    }

    for kind, namedInstances := range c.Functionality {

        if len(*namedInstances) == 0 {
            return fmt.Errorf("unused functionality type %v", kind)
        }

        err = namedInstances.setType(kind)
        if err != nil {
            return err
        }

        err = namedInstances.setName()

        err = namedInstances.check()
        if err != nil {
            return err
        }
    }
    
    if len(c.DataPaths) == 0 {
        return fmt.Errorf("you must declare data paths in a capsule node")
    }

    for path, settings := range c.DataPaths {
        if settings == nil {
            return fmt.Errorf("remove empty data path: %v", path)
        }

        err = settings.check()
        if err != nil {
            return fmt.Errorf("error in data path %v: %v", path, err)
        }
    }

    if len(c.Ingest) == 0 {
        return fmt.Errorf("aegis config must specify the vlans of the attached nodes under ingest")
    }

    for node, ingest := range c.Ingest {
        err = ingest.check()
        if err != nil {
            return fmt.Errorf("error in ingest data for node %v: %v", node, err)
        }
    }

    if c.Debug == nil { c.Debug = &DebugInfo{} }

    return nil
}


func (n *NamedFunctionality) setType(kind FunctionType) error {
    if !kind.isValid() {  
        return fmt.Errorf("function type %v is invalid", kind)
    }

    if len(*n) == 0 {
        return fmt.Errorf("remove empty function type declaration %v", kind)
    }

    if kind == Silent {
        for name, _ := range *n {
            (*n)[name] = &FunctionalityInfo{}
        }
    }

    for name, settings := range *n {
        if settings == nil {
            return fmt.Errorf("remove empty function %v", name)
        }

        if kind == Silent { }

        settings.Type = kind
    }

    return nil
}


func (n *NamedFunctionality) setName() error {
    if len(*n) == 0 {
        return fmt.Errorf("aegis nodes must declare functionality")
    }

    for name, settings := range *n {
        settings.Name = name
    }

    return nil
}


func (n *NamedFunctionality) check() error {
    if len(*n) == 0 {
        return fmt.Errorf("aegis nodes must declare functionality")
    }

    for name, info := range *n {
        if info == nil && info.Type != Silent {
            return fmt.Errorf("remove empty function %v", name)
        }

        if info == nil { continue }

        err := info.check()
        if err != nil {
            return fmt.Errorf("error verifying settings for %v: err", name, err)
        }
    }
    return nil
}


func (d *DataPath) check() error {
    if d.Type == "" { d.Type = Bidirectional; }

    if !d.Type.isValid() {
        return fmt.Errorf("invalid data path type. Only valid types are unidirection and bidirectional")
    }

    // Add some error catching to verify these are all valid nodes and functions. Skipping for now

    return nil
}


func (i *IngestInfo) check() error {
    if i.Vlan == 0 {
        return fmt.Errorf("the vlan id of every external node must be specified")
    }

    return nil
}


func (f *FunctionalityInfo) check() error {
    var err error;

    if f.Name == "" {
        return fmt.Errorf("all functionality must be named")
    }

    if !f.Type.isValid() {
        return fmt.Errorf("invalid functionality type %v", f.Type)
    }

    for _, route := range f.Routes {
        err = route.check()
        if err != nil {
            return fmt.Errorf("invalid route in function %v. err: %v", f.Name, err)
        }
    }

    for _, ethX := range f.Interfaces {
        err = ethX.check()
        if err != nil {
            return fmt.Errorf("invalid interface in function %v. err: %v", f.Name, err)
        }
    }

    if f.Type == IpTables {
        return f.checkIpTables()
    }

    if f.Type == WireGuard {
        return f.checkWireGuard()
    }

    return nil
}


func (f *FunctionalityInfo) checkIpTables() error {
    if len(f.Setup) == 0 {
        return fmt.Errorf("firewall %v has no setup script", f.Name)
    }

    return nil
}


func (f *FunctionalityInfo) checkWireGuard() error {
    var err error

    // Verify the ingest address
    if f.IngestAddr == "" {
        return fmt.Errorf("no ingest address set for wireguard host %v", f.Name)
    }

    _, _, err = net.ParseCIDR(f.IngestAddr)
    if err != nil {
        return fmt.Errorf("invalid ingest address for wireguard host %v. err: %v", f.Name, err)
    }

    // Verify the egress address
    if f.EgressAddr == "" {
        return fmt.Errorf("no egress address set for wireguard host %v", f.Name)
    }

    _, _, err = net.ParseCIDR(f.EgressAddr)
    if err != nil {
        return fmt.Errorf("invalid egress address for wireguard host %v. err: %v", f.Name, err)
    }

    // Verify the Wire Guard interface
    err = f.WgInterface.check()
    if err != nil {
        return fmt.Errorf("invalid wireguard interface config for host %v. err: %v", f.Name, err)
    }

    // Verify the peers are correct
    if len(f.Peers) == 0 {
        return fmt.Errorf("no peers specified for wireguard host %v", f.Name)
    }

    for name, peer := range f.Peers {
        if name == "" {
            return fmt.Errorf("all wireguard peers must be name in wireguard host %v", f.Name)
        }

        err = peer.check()
        if err != nil {
            return fmt.Errorf("error building peer %v on host %v. err %v", name, f.Name, err)
        }
    }

    return nil
}


func (w WgInterface) check() error {
    if w.Name == "" {
        return fmt.Errorf("wireguard interfaces is unnamed")
    }

    if w.PrivateKey == "" {
        return fmt.Errorf("wireguard interface doesn't specify a private key")
    }

    _, _, err := net.ParseCIDR(w.Addr)
    if err != nil {
        return fmt.Errorf("invalid address for wireguard interface %v", w.Addr)
    }

    if w.Port == 0 {
        return fmt.Errorf("wireguard interface doesn't specify a port")
    }

    if w.Port < 1 || w.Port > 65535 {
        return fmt.Errorf("port number %v out of range", w.Port)
    }

    return nil
}


func (p WgPeer) check() error {
    if p.PublicKey == "" {
        return fmt.Errorf("wireguard peer doesn't specify a public key")
    }

    // Verify the endpoint
    host, port, err := net.SplitHostPort(p.Endpoint)
    if err != nil {
        return err
    }

    if net.ParseIP(host) == nil {
        return fmt.Errorf("invalid endpoint ip for wireguard peer")
    }

    portNum, err := strconv.Atoi(port)
    if err != nil {
        return err
    } else if portNum < 1 || portNum > 65535 {
        return fmt.Errorf("port number %v out of range", portNum)
    }

    // Verify the allowed IPs
    for _, ip := range p.AllowedIPs {
        _, _, err = net.ParseCIDR(ip)
        if err != nil {
            return fmt.Errorf("invalid allowed ip for wireguard peer %v", p.Endpoint)
        }
    }

    return nil
}


func (r ForwardingPath) check() error {
    _, _, err := net.ParseCIDR(r.To)
    if err != nil {
        return fmt.Errorf("invalid to ip %v", r.To)
    }

    isValid := net.ParseIP(r.Via) != nil
    if !isValid {
        return fmt.Errorf("invalid via ip %v", r.Via)
    }

    return nil
}


func (i Interface) check() error {
    if i.Dev == "" {
        return fmt.Errorf("interface must have device name")
    }

    _, _, err := net.ParseCIDR(i.Addr)
    if err != nil {
        return fmt.Errorf("invalid via ip %v", i.Addr)
    }

    return nil
}
