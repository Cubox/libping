/*
This is a munin plugin.
For adding hosts to ping, add in conf:
[ping]
user root
group root
env.hosts ping.me and.me
*/

package main

import (
    "fmt"
    "github.com/Cubox-/libping"
    "os"
    "strings"
)

func main() {
    if os.Getenv("hosts") == "" {
        os.Exit(1)
    }

    if len(os.Args) > 1 && os.Args[1] == "config" {
        for _, host := range strings.Split(os.Getenv("hosts"), " ") {
            fmt.Printf("multigraph ping_%s\n", strings.Replace(host, ".", "_", -1))
            fmt.Printf("graph_title Latency to %s\n", host)
            fmt.Println("graph_category latency")
            fmt.Println("graph_vlabel ping")
            fmt.Println("graph_scale no")
            fmt.Printf("graph_info This graph show the latency to reach %s\n", host)
            fmt.Println("ping.label ping")
        }
    } else if len(os.Args) == 1 {
        for _, host := range strings.Split(os.Getenv("hosts"), " ") {
            libping.Pingonce(host)
            duration, err := libping.Pingonce(host)

            if err != nil {
                continue
            }

            fmt.Printf("multigraph ping_%s\n", strings.Replace(host, ".", "_", -1))
            fmt.Printf("ping.value %.5f\n", float64(duration.Nanoseconds())/float64(1000000))
        }
    } else {
        os.Exit(1)
    }
}
