package main

import (
    "fmt"
    ping "github.com/Cubox-/libping"
    "time"
)

func main() {
    chann := make(chan ping.Response, 100)
    go ping.Pinguntil("google.com", 0, chann, time.Second)
    for i := range chann {
        fmt.Println(i)
    }
}
