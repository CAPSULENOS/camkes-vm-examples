package main

import (
    "os"
    "fmt"
    "time"
    "orchestration/helpers"
    "orchestration/takedown"
    "orchestration/types"
    "orchestration/functionality"
    "orchestration/nodes"
    "orchestration/build"
)

type FunctionalityInfo = types.FunctionalityInfo
type NodeInfo = types.NodeInfo
type Config  = types.Config


func main() {

    if len(os.Args[1:]) == 0 {
        helpMenu()
        return
    }

    config_file := "/etc/aegis/aegis-conf.yml"
    lease_file := "/var/run/dnsmasq-br0.leases"

    switch os.Args[1] {
        case "start":
            StartAegis(config_file, lease_file)

        case "stop":
            takedown.StopAegis(lease_file)

        default: 
            helpMenu()
    }

}

func helpMenu() {
    fmt.Println("aegis [options]")
    fmt.Println("  start   Start Aegis")
    fmt.Println("  stop    Stop Aegis")
}

func StartAegis(config_file string, lease_file string) {

    // Read the config file
    var network_settings Config
    var err error

    network_settings, err = helpers.LoadSettings(config_file)
    if err != nil {
        helpers.LogE(err);
        return
    }

    // Initialize the function_settings information
    function_settings := functionality.PopulateFunctionSettings(network_settings)
    

    // Get physical nodes to allocate VNF functionality to
    var available_nodes []NodeInfo


    available_nodes, err = nodes.GetAvailableNodes(lease_file)
    for len(available_nodes) < network_settings.NumberOfVms || err != nil {
        if err != nil {
            helpers.LogE("Error finding the number of connected nodes.", err, "Exiting")
            return
        }
        
        helpers.LogE("No available VMs on the network. Waiting for VMs...")
        time.Sleep(time.Second)
        
        // Update available_nodes and err for the next iteration
        available_nodes, err = nodes.GetAvailableNodes(lease_file)
    }

    // Set up /etc/hosts for ease of use
    err = build.EtcHosts(available_nodes)
    if err != nil {
        helpers.LogE("Error updaing /etc/hosts.", err, "Exiting")
        return
    }


    // Give each function a unique vlan tag in and out
    vid_index := 2; // 1 is reserved

    for name, datapath_settings := range network_settings.DataPaths {
        helpers.LogE("Setting up Data Path:", name)

        // Set up ingest data:
        datapath_settings.Ingest = network_settings.Ingest[datapath_settings.Path[0]]

        // Add the ingest connection 
        datapath_settings.Connections = [][2]int{[2]int{datapath_settings.Ingest.Vlan, vid_index}}

        last_vid := 0;  // Used to make internal connections. init @ Ingest so we auto get that connection

        // Build the forward data path
        forwardInternalPath := datapath_settings.Path[1:]
        function_settings, datapath_settings, vid_index, last_vid, err = functionality.FinalizeInternalForwardPath(forwardInternalPath, function_settings, datapath_settings, available_nodes, vid_index, last_vid)
        if err != nil {
            helpers.LogE("Error building internal forward data path", err)
        }


        // Save those changes
        network_settings.DataPaths[name] = datapath_settings

        // Should we set up the return path?
        if datapath_settings.Type != "bidirectional" { continue }

        // We increment by 1 extra in previous loop. Use out in forward as in in reverse
        last_vid -= 1
        vid_index -= 1

        // Save those settings
        network_settings.DataPaths[name] = datapath_settings
    }


    // Actually build the ovs forwarding topology
    for _, settings := range network_settings.DataPaths {
        err = build.ForwardingPath(settings.Connections, settings.Type)
        if err != nil {
            helpers.LogE(err)
            return
        }
    }
    // Build out the VNFs on sel4 virtual machines
    for name, settings := range function_settings {
        err = build.FunctionsOnNode(name, settings)
        if err != nil {
            helpers.LogE(err)
            return
        }
    }
}

