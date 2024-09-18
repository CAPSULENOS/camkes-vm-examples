package main

import (
    "os"
    "fmt"

    log "github.com/sirupsen/logrus"

    "github.com/CAPSULENOS/aegis/pkg/config"
)

var (
    verbose bool
)

func main() {
    if len(os.Args[1:]) == 0 {
        helpMenu()
        return
    }

    switch os.Args[1] {
        case "start":
            doStart()

        case "stop":
            doStop()

        default: 
            helpMenu()
    }
}

func doStart() {
    cfg := &config.Config{}

    err := cfg.LoadFromFile("/etc/aegis/aegis-conf.yml")
    if err != nil {
        log.Fatal(err)
    }

    for _, f := range cfg.Functionality {
        for n, i := range *f {
            fmt.Println(n, i)
        }
    }
}

func doStop() {
    fmt.Println("To do")
}

func helpMenu() {
    fmt.Println("aegis [options]")
    fmt.Println("  start   Start Aegis")
    fmt.Println("  stop    Stop Aegis")
}
