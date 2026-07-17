package gwlb

import (
	"context"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"net"
	"net/http"
)

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	intendedAddr := ctx.Value(http.LocalAddrContextKey).(*net.TCPAddr)
	sourceAddr, _ := AddrsFromContext(ctx)
	netstack := NetstackFromContext(ctx)

	remote := tcpip.FullAddress{Addr: tcpip.AddrFrom4Slice(intendedAddr.IP.To4()), Port: uint16(intendedAddr.Port)}
	local := tcpip.FullAddress{Addr: tcpip.AddrFrom4Slice(sourceAddr.IP.To4()), Port: uint16(sourceAddr.Port)}
	return gonet.DialTCPWithBind(ctx, netstack, local, remote, ipv4.ProtocolNumber)
}
