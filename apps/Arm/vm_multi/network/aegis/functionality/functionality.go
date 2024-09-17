package functionality

import (
    "orchestration/types"
)

type Config = types.Config
type FunctionalityInfo = types.FunctionalityInfo
type NodeInfo = types.NodeInfo
type DebugInfo  = types.DebugInfo
type DataPath = types.DataPath
type IngestInfo = types.IngestInfo


func PopulateFunctionSettings(network_settings Config) (map[string]FunctionalityInfo) {
    function_settings :=  make(map[string]FunctionalityInfo)

    for kind, inst := range network_settings.Functionality {
        for name, details := range inst {
            details.Type = kind
            function_settings[name] = details
        }
    }

    return function_settings
}


func FinalizeFunctionSettings(function_map map[string]FunctionalityInfo, function_name string, ingest_node_name string, available_nodes []NodeInfo, new_vids [2]int) map[string]FunctionalityInfo {

    settings := function_map[function_name]

    switch settings.Type {
        case "router":
            for i, interfaces := range settings.Interfaces {
                if interfaces.Dev != ingest_node_name { continue }
                settings.Interfaces[i].Vlan = new_vids[0]
            }
        case "silent":
            if len(settings.Vlans) == 0 { 
                settings.Vlans = [][2]int{ new_vids } 
            } else {
                settings.Vlans = append(settings.Vlans, new_vids)
            }
        case "wireguard":
            if len(settings.Vlans) == 0 { 
                settings.Vlans = [][2]int{ new_vids } 
            } else {
                settings.Vlans = append(settings.Vlans, new_vids)
            }
        default:
            return function_map
    }

    if settings.PhysicalNode == nil {
        settings = AssignFunctionalityToInfraNode(settings, available_nodes)
    }

    function_map[function_name] = settings

    return function_map
}

var counter int;

func AssignFunctionalityToInfraNode(function_settings FunctionalityInfo, available_nodes []NodeInfo) FunctionalityInfo {

    function_settings.PhysicalNode = &available_nodes[counter];

    counter = (counter + 1) % len(available_nodes);

    return function_settings
}


// Below is where the data path is build and various fields are populated


// Wrapper so the main statement is easier to read but we don't repeat ourselves
func FinalizeInternalForwardPath(internal_path []string, function_settings map[string]FunctionalityInfo, datapath_settings DataPath, available_nodes []NodeInfo, vid_index int, last_vid int) (map[string]FunctionalityInfo, DataPath, int, int, error) {
    return populateFieldsNecessaryToBuildInternalDataPath(internal_path, function_settings, datapath_settings, available_nodes, vid_index, last_vid)
}

func FinalizeInternalReversePath(internal_path []string, function_settings map[string]FunctionalityInfo, datapath_settings DataPath, available_nodes []NodeInfo, vid_index int, last_vid int) (map[string]FunctionalityInfo, DataPath, int, int, error) {
    return populateFieldsNecessaryToBuildInternalDataPath(internal_path, function_settings, datapath_settings, available_nodes, vid_index, last_vid)
}


type BuildInternalPathError struct {
    s string
}

func (e *BuildInternalPathError) Error() string {
    return e.s
}

func NewBuildInternalPathError(message string) *BuildInternalPathError {
    return &BuildInternalPathError{s: message}
}

func populateFieldsNecessaryToBuildInternalDataPath(internal_path []string, function_settings map[string]FunctionalityInfo, datapath_settings DataPath, available_nodes []NodeInfo, vid_index int, last_vid int) (map[string]FunctionalityInfo, DataPath, int, int, error) {
    for _, function_name := range internal_path {

        if vid_index + 1 >= 4096 { 
            err := NewBuildInternalPathError("Too many vlans required for this configuration")
            return function_settings, datapath_settings, vid_index, last_vid, err
        }

        // Assign VIDs to a given function
        new_vids := [2]int{vid_index, vid_index + 1}

        function_settings = FinalizeFunctionSettings(function_settings, function_name, datapath_settings.Path[0], available_nodes, new_vids)

        if last_vid > 0 {
            datapath_settings.Connections =  append(datapath_settings.Connections, [2]int{last_vid, vid_index})
        }

        last_vid = vid_index + 1 // Outbound VID path
        // increment vid_index to prevent overlap
        vid_index += 2
    }

    return function_settings, datapath_settings, vid_index, last_vid, nil
}

