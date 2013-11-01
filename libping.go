/*
Package libping provide the ability to send ICMP packets easily.
*/
package libping

import (
    "bytes"
    "net"
    "os"
    "time"
)

const (
    ICMP_ECHO_REQUEST = 8
    ICMP_ECHO_REPLY   = 0
)

// The struct Response is the data returned by Pinguntil.
type Response struct {
    Delay       time.Duration
    Error       error
    Destination string
}

func makePingRequest(id, seq, pktlen int, filler []byte) []byte {
    p := make([]byte, pktlen)
    copy(p[8:], bytes.Repeat(filler, (pktlen-8)/len(filler)+1))

    p[0] = ICMP_ECHO_REQUEST // type
    p[1] = 0                 // code
    p[2] = 0                 // cksum
    p[3] = 0                 // cksum
    p[4] = uint8(id >> 8)    // id
    p[5] = uint8(id & 0xff)  // id
    p[6] = uint8(seq >> 8)   // sequence
    p[7] = uint8(seq & 0xff) // sequence

    // calculate icmp checksum
    cklen := len(p)
    s := uint32(0)
    for i := 0; i < (cklen - 1); i += 2 {
        s += uint32(p[i+1])<<8 | uint32(p[i])
    }
    if cklen&1 == 1 {
        s += uint32(p[cklen-1])
    }
    s = (s >> 16) + (s & 0xffff)
    s = s + (s >> 16)

    // place checksum back in header; using ^= avoids the
    // assumption the checksum bytes are zero
    p[2] ^= uint8(^s & 0xff)
    p[3] ^= uint8(^s >> 8)

    return p
}

func parsePingReply(p []byte) (id, seq int) {
    id = int(p[4])<<8 | int(p[5])
    seq = int(p[6])<<8 | int(p[7])
    return
}

// Pingone send one ICMP echo packet to the destination, and return the latency.
func Pingone(destination string) (time.Duration, error) {
    raddr, err := net.ResolveIPAddr("ip", destination)
    if err != nil {
        return 0, err
    }

    ipconn, err := net.Dial("ip:icmp", raddr.IP.String())

    if err != nil {
        return 0, err
    }

    defer ipconn.Close()

    sendid := os.Getpid() & 0xffff
    pingpktlen := 64

    sendpkt := makePingRequest(sendid, 0, pingpktlen, []byte("Go Ping"))

    start := time.Now()
    n, err := ipconn.Write(sendpkt)

    if err != nil || n != pingpktlen {
        return 0, err
    }

    ipconn.SetReadDeadline(time.Now().Add(time.Second * 1))

    resp := make([]byte, 1024)
    for {
        _, err := ipconn.Read(resp)

        if resp[1] != ICMP_ECHO_REPLY {
            continue
        } else if err != nil {
            return 0, err
        } else {
            return time.Now().Sub(start), nil
        }
    }
}

// Pinguntil will send ICMP echo packets to the destination until the counter is done, or forever if the counter is set to 0.
// The replies are given in the Response format, with the latency and the error, if one was got.
// You can also adjust the delay between two ICMP echo packets with the variable delay
func Pinguntil(destination string, count int, response chan Response, delay time.Duration) {
    raddr, err := net.ResolveIPAddr("ip", destination)
    if err != nil {
        response <- Response{Delay: 0, Error: err, Destination: destination}
        return
    }

    ipconn, err := net.Dial("ip:icmp", raddr.IP.String())
    if err != nil {
        response <- Response{Delay: 0, Error: err, Destination: destination}
        return
    }

    sendid := os.Getpid() & 0xffff
    pingpktlen := 64
    seq := 0

    for ; seq < count || count == 0; seq++ {
        sendpkt := makePingRequest(sendid, seq, pingpktlen, []byte("Go Ping"))

        start := time.Now()

        n, err := ipconn.Write(sendpkt)
        if err != nil || n != pingpktlen {
            response <- Response{Delay: 0, Error: err, Destination: destination}
            time.Sleep(delay)
            continue
        }

        ipconn.SetReadDeadline(time.Now().Add(time.Second * 1)) // 1 second

        resp := make([]byte, 1024)
        for {
            _, err := ipconn.Read(resp)

            if resp[1] != ICMP_ECHO_REPLY {
                continue
            } else if err != nil {
                response <- Response{Delay: 0, Error: err, Destination: destination}
                break
            } else {
                response <- Response{Delay: time.Now().Sub(start), Error: err, Destination: destination}
                break
            }
        }
        time.Sleep(delay)
    }
    close(response)
}