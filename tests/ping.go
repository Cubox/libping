package main

import (
    "fmt"
    ping "github.com/Cubox-/libping"
    "net"
    "os"
    "time"
)

func main() {
    chann := make(chan ping.Response, 100)
    go ping.Pinguntil(os.Args[1], 0, chann, time.Second)
    for i := range chann {
        if ne, ok := i.Error.(net.Error); ok && ne.Timeout() {
            fmt.Printf("Request timeout for icmp_seq %d\n", i.Seq)
            continue
        } else if i.Error != nil {
            fmt.Println(i.Error)
        } else {
            fmt.Printf("%d bytes from %s: icmp_seq=%d time=%s\n", i.Readsize, i.Destination, i.Seq, i.Delay)
        }
    }
}
