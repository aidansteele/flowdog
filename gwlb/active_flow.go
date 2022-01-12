package gwlb

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aidansteele/flowdog/mytls"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"inet.af/netstack/tcpip"
	"inet.af/netstack/tcpip/adapters/gonet"
	"inet.af/netstack/tcpip/buffer"
	"inet.af/netstack/tcpip/header"
	"inet.af/netstack/tcpip/link/channel"
	"inet.af/netstack/tcpip/network/ipv4"
	"inet.af/netstack/tcpip/stack"
	"inet.af/netstack/tcpip/transport/tcp"
	"net"
	"net/http"
	"time"
)

const (
	timeoutTcp    = 350 * time.Second
	timeoutNonTcp = 120 * time.Second
)

var errTimeout = errors.New("gwlb timeout")

type activeFlow struct {
	geneveHeader []byte
	gwlbConn     *net.UDPConn
	endpoint     *channel.Endpoint
	stack        *stack.Stack
	httpReady    chan struct{}
	handler      http.Handler
}

type FlowAcceptor interface {
	AcceptFlow(ctx context.Context, pkt gopacket.Packet, opts AwsGeneveOptions) bool
}

func newFlow(ctx context.Context, ch chan genevePacket, acceptor FlowAcceptor, opts AwsGeneveOptions, handler http.Handler) {
	ctx = ContextWithGeneveOptions(ctx, opts)

	// retrieve first packet in flow to inspect ip, port, etc
	pkt := <-ch
	// then we reinject for forwarding/interception/whatever
	go func() { ch <- pkt }()

	if acceptor != nil && !acceptor.AcceptFlow(ctx, pkt.pkt, opts) {
		return // TODO: this should probably send a refusal (for tcp) instead of silent dropping
	}

	fmt.Printf("new flow vpcEndpointId=%016x attachmentId=%016x flowCookie=%08x\n", opts.VpcEndpointId, opts.AttachmentId, opts.FlowCookie)
	gwlbConn := getUdpConn(pkt.addr)

	ipLayer, isIpv4 := pkt.pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	tcpLayer, isTcp := pkt.pkt.TransportLayer().(*layers.TCP)

	if !isTcp {
		fmt.Printf("fast-path for non-tcp flow=%08x\n", opts.FlowCookie)
		fastPath(ctx, ch, gwlbConn, timeoutNonTcp)
		return
	}

	if !isIpv4 {
		fmt.Printf("fast-path for non-ipv4 flow=%08x\n", opts.FlowCookie)
		fastPath(ctx, ch, gwlbConn, timeoutTcp)
		return
	}

	if handler == nil || (tcpLayer.DstPort != 80 && tcpLayer.DstPort != 443) {
		fmt.Printf("fast-path for non-port 80/443 flow=%08x\n", opts.FlowCookie)
		fastPath(ctx, ch, gwlbConn, timeoutTcp)
		return
	}

	geneveLayer := pkt.pkt.Layer(layers.LayerTypeGeneve).(*layers.Geneve)
	contents := geneveLayer.LayerContents()
	hdr := make([]byte, len(contents))
	copy(hdr, contents)

	sourceAddr := &net.TCPAddr{IP: ipLayer.SrcIP, Port: int(tcpLayer.SrcPort)}
	ctx = ContextWithSourceAddr(ctx, sourceAddr)

	endpoint, netstack := newEndpointAndStack(opts)
	ctx = ContextWithNetstack(ctx, netstack)

	a := &activeFlow{
		geneveHeader: hdr,
		gwlbConn:     gwlbConn,
		endpoint:     endpoint,
		stack:        netstack,
		httpReady:    make(chan struct{}, 1),
		handler:      handler,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.runNetstackToGwlb(ctx)
	})

	g.Go(func() error {
		return a.runProxyHttp(ctx)
	})

	g.Go(func() error {
		<-a.httpReady
		return a.runGwlbToNetstack(ctx, ch, timeoutTcp)
	})

	err := g.Wait()
	if err == errTimeout {
		return
	}

	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
}

func (a *activeFlow) runGwlbToNetstack(ctx context.Context, ch chan genevePacket, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout):
			return errTimeout
		case pkt := <-ch:
			geneveLayer := pkt.pkt.Layer(layers.LayerTypeGeneve).(*layers.Geneve)
			payload := geneveLayer.LayerPayload()
			data := buffer.NewVectorisedView(1, []buffer.View{payload})
			newPkt := stack.NewPacketBuffer(stack.PacketBufferOptions{Data: data})
			a.endpoint.InjectInbound(ipv4.ProtocolNumber, newPkt)
			newPkt.DecRef()
		}
	}
}

func (a *activeFlow) runNetstackToGwlb(ctx context.Context) error {
	for {
		pinfo, more := a.endpoint.ReadContext(ctx)
		if !more {
			return ctx.Err()
		}

		data := buffer.NewVectorisedView(pinfo.Pkt.Size(), pinfo.Pkt.Views())
		view := data.ToView()

		err := a.sendGeneve(view)
		if err != nil {
			return err
		}
	}
}

func (a *activeFlow) runProxyHttp(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	httpSrv := &http.Server{
		Handler:     a.handler,
		TLSConfig:   mytls.TlsConfig(),
		BaseContext: func(listener net.Listener) context.Context { return ctx },
	}

	httpSrv.TLSConfig.NextProtos = []string{"h2", "http/1.1"}

	port443, err := gonet.ListenTCP(a.stack, tcpip.FullAddress{Port: 443}, ipv4.ProtocolNumber)
	if err != nil {
		return errors.WithStack(err)
	}

	port80, err := gonet.ListenTCP(a.stack, tcpip.FullAddress{Port: 80}, ipv4.ProtocolNumber)
	if err != nil {
		return errors.WithStack(err)
	}

	a.httpReady <- struct{}{}

	g.Go(func() error {
		err := httpSrv.Serve(port80)
		return errors.WithStack(err)
	})

	g.Go(func() error {
		err := httpSrv.ServeTLS(port443, "", "")
		return errors.WithStack(err)
	})

	g.Go(func() error {
		<-ctx.Done()
		return httpSrv.Close()
	})

	return g.Wait()
}

func (a *activeFlow) sendGeneve(body []byte) error {
	buf := &bytes.Buffer{}
	buf.Write(a.geneveHeader)
	buf.Write(body)

	_, err := a.gwlbConn.Write(buf.Bytes())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func newEndpointAndStack(opts AwsGeneveOptions) (*channel.Endpoint, *stack.Stack) {
	linkAddr, _ := tcpip.ParseMACAddress("aa:bb:cc:dd:ee:ff")
	endpoint := channel.New(200, 1500, linkAddr)

	s := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	// Add default route.
	s.SetRouteTable([]tcpip.Route{
		{NIC: NICID, Destination: header.IPv4EmptySubnet},
	})

	linkEndpoint := endpoint
	//linkEndpoint := sniffer.NewWithPrefix(endpoint, fmt.Sprintf("flow-%08x ", opts.flowCookie))
	tcpErr := s.CreateNIC(NICID, linkEndpoint)
	if tcpErr != nil {
		fmt.Printf("%+v\n", tcpErr)
		panic(tcpErr)
	}

	s.SetPromiscuousMode(NICID, true)
	s.SetSpoofing(NICID, true)
	return endpoint, s
}

const NICID = 1