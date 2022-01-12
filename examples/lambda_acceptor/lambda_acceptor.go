package lambda_acceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
)

type LambdaAcceptor struct {
	lambda   lambdaiface.LambdaAPI
	function string
	enricher *Enricher
}

func New(lambda lambdaiface.LambdaAPI, function string, ec2 ec2iface.EC2API) (*LambdaAcceptor, error) {
	enricher, err := NewEnricher(ec2)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &LambdaAcceptor{
		lambda:   lambda,
		function: function,
		enricher: enricher,
	}, nil
}

type AcceptorInput struct {
	GeneveOptions     GeneveOptions
	NetworkProtocol   string
	TransportProtocol string
	Source            Endpoint
	Destination       Endpoint
}

type GeneveOptions struct {
	VpcEndpointId string
	FlowCookie    string
	AttachmentId  string `json:"AttachmentId,omitempty"`
}

type Endpoint struct {
	IpAddress string
	Port      int          `json:"Port,omitempty"`
	EC2       *Ec2Endpoint `json:"EC2,omitempty"`
}

type Ec2Endpoint struct {
	InstanceId string
	AccountId  string
	VpcId      string
	SubnetId   string
	Tags       map[string]string
}

type AcceptorOutput struct {
	Accept bool
}

func (a *LambdaAcceptor) AcceptFlow(ctx context.Context, pkt gopacket.Packet, opts gwlb.AwsGeneveOptions) bool {
	sourceIp := ""
	destIp := ""
	networkProto := ""

	switch network := pkt.NetworkLayer().(type) {
	case *layers.IPv4:
		networkProto = "IPv4"
		sourceIp = network.SrcIP.String()
		destIp = network.DstIP.String()
	case *layers.IPv6:
		networkProto = "IPv6"
		sourceIp = network.SrcIP.String()
		destIp = network.DstIP.String()
	}

	sourcePort := 0
	destPort := 0
	transportProto := ""

	switch transport := pkt.TransportLayer().(type) {
	case *layers.TCP:
		transportProto = "TCP"
		sourcePort = int(transport.SrcPort)
		destPort = int(transport.DstPort)
	case *layers.UDP:
		transportProto = "UDP"
		sourcePort = int(transport.SrcPort)
		destPort = int(transport.DstPort)
	}

	attachmentId := ""
	if opts.AttachmentId > 0 {
		attachmentId = fmt.Sprintf("%016x", opts.AttachmentId)
	}

	inputPayload, _ := json.Marshal(AcceptorInput{
		GeneveOptions: GeneveOptions{
			VpcEndpointId: fmt.Sprintf("vpce-0%016x", opts.VpcEndpointId),
			FlowCookie:    fmt.Sprintf("%08x", opts.FlowCookie),
			AttachmentId:  attachmentId,
		},
		NetworkProtocol:   networkProto,
		TransportProtocol: transportProto,
		Source: Endpoint{
			IpAddress: sourceIp,
			Port:      sourcePort,
			EC2:       a.ec2Endpoint(sourceIp),
		},
		Destination: Endpoint{
			IpAddress: destIp,
			Port:      destPort,
			EC2:       a.ec2Endpoint(destIp),
		},
	})

	invoke, err := a.lambda.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName: &a.function,
		Payload:      inputPayload,
	})
	if err != nil {
		fmt.Printf("lambda acceptor err %+v\n", err)
		return false
	}

	output := AcceptorOutput{}
	err = json.Unmarshal(invoke.Payload, &output)
	if err != nil {
		fmt.Printf("parsing lambda acceptor output err %+v\n", err)
		return false
	}

	return output.Accept
}

// TODO: this enriched info should be part of the ctx
// earlier on so it can be used by interceptors etc
func (a *LambdaAcceptor) ec2Endpoint(ip string) *Ec2Endpoint {
	instance := a.enricher.InstanceByIp(ip)
	if instance == nil {
		return nil
	}

	iface := instance.NetworkInterfaces[0]
	tags := map[string]string{}
	for _, tag := range instance.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return &Ec2Endpoint{
		InstanceId: *instance.InstanceId,
		AccountId:  *iface.OwnerId,
		VpcId:      *iface.VpcId,
		SubnetId:   *iface.SubnetId,
		Tags:       tags,
	}
}
