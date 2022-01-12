package gwlb

import (
	"context"
	"fmt"
	"net"
	"time"
)

func fastPath(ctx context.Context, ch chan genevePacket, gwlbConn *net.UDPConn, timeout time.Duration) {
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
		case <-time.After(timeout):
			return
		}
	}
}
