package gwlb

import (
	"fmt"
	"github.com/tidwall/spinlock"
	"net"
)

var udpLock spinlock.Locker
var udpMap map[int]*net.UDPConn

func getUdpConn(remote *net.UDPAddr) *net.UDPConn {
	udpLock.Lock()
	defer udpLock.Unlock()

	lport := remote.Port
	if udpMap == nil {
		udpMap = map[int]*net.UDPConn{}
	}

	gwlbConn := udpMap[lport]
	if gwlbConn != nil {
		fmt.Println("reusing udp conn")
		return gwlbConn
	}

	// TODO: gwlb will have a different IP per AZ, right? is hard-coding it here a problem?
	raddr := &net.UDPAddr{IP: remote.IP, Port: 6081}
	laddr := &net.UDPAddr{Port: lport}

	var err error
	gwlbConn, err = net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	udpMap[lport] = gwlbConn
	return gwlbConn
}
