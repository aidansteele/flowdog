package gwlb

import (
	"context"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb/shark"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

type genevePacket struct {
	buf  []byte
	addr *net.UDPAddr
	pkt  gopacket.Packet
}

type Server struct {
	Acceptor FlowAcceptor
	Handler  http.Handler
}

func (s *Server) Serve(ctx context.Context, conn *net.UDPConn) error {
	flows := map[uint32]chan genevePacket{}
	flowEndedCh := make(chan uint32)

	pktbuf := make([]byte, 16_384)

	sharkch := make(chan []byte)
	ss := shark.NewSharkServer()

	sharkL, err := net.Listen("tcp", "127.0.0.1:7081")
	if err != nil {
		return errors.WithStack(err)
	}
	go ss.Serve(ctx, sharkL, sharkch)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case flowCookie := <-flowEndedCh:
			fmt.Printf("deleting flow %08x\n", flowCookie)
			delete(flows, flowCookie)
		default:
			n, addr, err := conn.ReadFromUDP(pktbuf)
			if err != nil {
				return errors.WithStack(err)
			}

			newbuf := make([]byte, n)
			copy(newbuf, pktbuf[:n])

			pkt := gopacket.NewPacket(newbuf, layers.LayerTypeGeneve, gopacket.Default)
			opts := ExtractAwsGeneveOptions(pkt)

			cookie := opts.FlowCookie
			ch, found := flows[cookie]
			if !found {
				ch = make(chan genevePacket)
				flows[cookie] = ch
				go func() {
					newFlow(ctx, ch, opts, newFlowOptions{
						acceptor:  s.Acceptor,
						handler:   s.Handler,
						mirror:    sharkch,
						keyLogger: ss.KeyLogWriter(),
					})
					flowEndedCh <- cookie
				}()
			}

			ch <- genevePacket{
				buf:  newbuf,
				addr: addr,
				pkt:  pkt,
			}
		}
	}
}
