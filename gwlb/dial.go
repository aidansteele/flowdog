package gwlb

import (
	"context"
	"inet.af/netstack/tcpip"
	"inet.af/netstack/tcpip/adapters/gonet"
	"inet.af/netstack/tcpip/network/ipv4"
	"net"
	"net/http"
)

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	intendedAddr := ctx.Value(http.LocalAddrContextKey).(*net.TCPAddr)
	sourceAddr := SourceAddrFromContext(ctx)
	netstack := NetstackFromContext(ctx)

	remote := tcpip.FullAddress{Addr: tcpip.Address(intendedAddr.IP.To4()), Port: uint16(intendedAddr.Port)}
	local := tcpip.FullAddress{Addr: tcpip.Address(sourceAddr.IP), Port: uint16(sourceAddr.Port)}
	return gonet.DialTCPWithBind(ctx, netstack, local, remote, ipv4.ProtocolNumber)
}
