package gwlb

import (
	"context"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb/mirror"
	"net"
	"time"
)

func fastPath(ctx context.Context, ch chan genevePacket, gwlbConn *net.UDPConn, mirrorch chan mirror.Packet, timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case pkt := <-ch:
			_, err := gwlbConn.Write(pkt.buf)
			if err != nil {
				fmt.Printf("%+v\n", err)
				panic(err)
			}
			mirrorch <- mirror.New(pkt.buf, true)
		case <-time.After(timeout):
			return
		}
	}
}
