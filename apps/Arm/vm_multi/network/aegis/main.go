package main

import (
    "time"
    "orchestration/helpers"
    "orchestration/types"
    "orchestration/functionality"
    "orchestration/nodes"
    "orchestration/build"
)

type FunctionalityInfo = types.FunctionalityInfo
type NodeInfo = types.NodeInfo
type Config  = types.Config


func main() {

    config_file := "/root/aegis-conf.yml"

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

    lease_file := "/var/run/dnsmasq-br0.leases"

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

        // Build the reverse data path
        reverseInternalPath := datapath_settings.Path[1:len(datapath_settings.Path) - 1]
        function_settings, datapath_settings, vid_index, last_vid, err = functionality.FinalizeInternalReversePath(reverseInternalPath, function_settings, datapath_settings, available_nodes, vid_index, last_vid)
        if err != nil {
            helpers.LogE("Error building internal reverse data path", err)
        }

        // Add the reverse egress path
        datapath_settings.Connections = append(datapath_settings.Connections, [2]int{last_vid, datapath_settings.Ingest.Vlan})
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
    
