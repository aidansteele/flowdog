package gwlb

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	AwsGeneveClass             = 0x108
	AwsGeneveTypeVpcEndpointId = 0x1
	AwsGeneveTypeAttachmentId  = 0x2
	AwsGeneveTypeFlowCookie    = 0x3
)

type AwsGeneveOptions struct {
	VpcEndpointId uint64
	AttachmentId  uint64
	FlowCookie    uint32
}

func ExtractAwsGeneveOptions(pkt gopacket.Packet) AwsGeneveOptions {
	geneveLayer := pkt.Layer(layers.LayerTypeGeneve).(*layers.Geneve)

	opts := AwsGeneveOptions{}

	for _, option := range geneveLayer.Options {
		if option.Class != AwsGeneveClass {
			continue
		}

		switch option.Type {
		case AwsGeneveTypeVpcEndpointId:
			opts.VpcEndpointId = binary.BigEndian.Uint64(option.Data)
		case AwsGeneveTypeAttachmentId:
			opts.AttachmentId = binary.BigEndian.Uint64(option.Data)
		case AwsGeneveTypeFlowCookie:
			opts.FlowCookie = binary.BigEndian.Uint32(option.Data)
		default:
			// no-op
		}
	}

	return opts
}
