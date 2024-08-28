package nodes

import (
    "os"
    "bufio"
    "strings"
    "orchestration/types"
)

type NodeInfo = types.NodeInfo


// Wrapper in case we want to add some logic here that isn't just return all connected nodes
func GetAvailableNodes(filename string) ([]NodeInfo, error) {
    connected, err := get_connected_nodes(filename)
    if err != nil {
        return nil, err
    }
    return connected, nil
}


func get_connected_nodes(filename string) ([]NodeInfo, error) {
    file, err := os.Open(filename)
    defer file.Close()
    if err != nil {
        return nil, err
    }

    // Create a scanner to read the file line by line
    scanner := bufio.NewScanner(file)

    var nodes []NodeInfo;

    // Loop through the file line by line
    for scanner.Scan() {
        parts := strings.Split(scanner.Text(), " ")
        nodes = append(nodes, NodeInfo{Sel4Name: parts[3], Mac: parts[1], Sel4IP:parts[2]})
    }

    // Check for errors during scanning
    if err := scanner.Err(); err != nil {
        return nil, err
    }

    // Print the number of lines
    return nodes, nil
}

