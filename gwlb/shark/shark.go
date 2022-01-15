package shark

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb/mirror"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net"
	"sync"
	"time"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative shark.proto

type SharkServer struct {
	keyLogWriter io.Writer
	keyLogReader io.Reader
	ctx          context.Context
	clients      []*client
	mut          sync.Mutex
}

func NewSharkServer() *SharkServer {
	pr, pw := io.Pipe()

	return &SharkServer{keyLogWriter: pw, keyLogReader: pr}
}

func (s *SharkServer) Serve(ctx context.Context, l net.Listener, ch chan mirror.Packet) error {
	s.ctx = ctx
	go s.run(ch)

	go s.streamSslKeyLog(ctx)

	gs := grpc.NewServer()
	RegisterVpcsharkServer(gs, s)

	go func() {
		<-ctx.Done()
		gs.Stop()
	}()

	err := gs.Serve(l)
	return errors.WithStack(err)
}

func (s *SharkServer) KeyLogWriter() io.Writer {
	return s.keyLogWriter
}

func (s *SharkServer) streamSslKeyLog(ctx context.Context) {
	scan := bufio.NewScanner(s.keyLogReader)
	buf := &bytes.Buffer{}

	for scan.Scan() {
		if ctx.Err() != nil {
			return
		}

		buf.Write(scan.Bytes())
		buf.WriteByte('\n')
		s.forEachClient(func(c *client) error {
			return c.stream.Send(&GetPacketsOutput{
				Time:      timestamppb.New(time.Now()),
				SslKeyLog: buf.Bytes(),
			})
		})
		buf.Reset()
	}
}

func (s *SharkServer) forEachClient(fn func(c *client) error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	validLen := 0

	for _, stream := range s.clients {
		err := fn(stream)
		if err == nil {
			s.clients[validLen] = stream
			validLen++
		}
	}

	// nil out invalid clients to garbage collect them
	for idx := validLen; idx < len(s.clients); idx++ {
		fmt.Println("removing an invalid shark listener")
		s.clients[idx] = nil
	}
	s.clients = s.clients[:validLen]
}

func (s *SharkServer) run(ch chan mirror.Packet) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case pkt := <-ch:
			geneve := gopacket.NewPacket(pkt.Packet, layers.LayerTypeGeneve, gopacket.Default).Layer(layers.LayerTypeGeneve).(*layers.Geneve)
			s.forEachClient(func(c *client) error {
				if c.mirrorType != pkt.Type {
					return nil
				}

				match := c.vm.Matches(gopacket.CaptureInfo{}, geneve.LayerPayload())
				if !match {
					return nil
				}

				return c.stream.Send(&GetPacketsOutput{
					Time:    timestamppb.New(time.Now()),
					Payload: pkt.Packet,
				})
			})
		}
	}
}

type client struct {
	stream     Vpcshark_GetPacketsServer
	vm         *pcap.BPF
	mirrorType mirror.Type
}

func (s *SharkServer) GetPackets(input *GetPacketsInput, stream Vpcshark_GetPacketsServer) error {
	vm, err := FilterVM(input.Filter)
	if err != nil {
		return err
	}

	mirrorType := mirror.TypeUnknown
	switch input.PacketType {
	case PacketType_PRE:
		mirrorType = mirror.TypePreRewrite
	case PacketType_POST:
		mirrorType = mirror.TypePostRewrite
	case PacketType_UNKNOWN:
		fallthrough
	default:
		return errors.New("unknown packet type")
	}

	s.mut.Lock()
	s.clients = append(s.clients, &client{stream: stream, vm: vm, mirrorType: mirrorType})
	s.mut.Unlock()

	<-s.ctx.Done()
	return nil
}

func (s *SharkServer) mustEmbedUnimplementedVpcsharkServer() {}
