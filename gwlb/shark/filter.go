package shark

import (
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
)

func FilterVM(input string) (*pcap.BPF, error) {
	vm, err := pcap.NewBPF(layers.LinkType(12), 0x10000, input)
	return vm, errors.WithStack(err)
}
