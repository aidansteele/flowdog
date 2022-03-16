package webfirehose

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aidansteele/flowdog/examples/account_id_emf"
	"github.com/aidansteele/flowdog/examples/lambda_acceptor"
	"github.com/aidansteele/flowdog/gwlb"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"time"
)

type requestRecord struct {
	TLS    *requestRecordTls `json:"TLS,omitempty"`
	AWS    *requestRecordAws `json:"AWS,omitempty"`
	Geneve requestRecordGeneve

	Time           time.Time
	Method         string
	Host           string
	ETldPlusOne    string
	URL            string
	RequestHeaders map[string]string
	Client         lambda_acceptor.Endpoint

	StatusCode      int
	StatusMessage   string
	ResponseHeaders map[string]string
	Server          lambda_acceptor.Endpoint
}

type requestRecordAws struct {
	AuthorizationHeader account_id_emf.ParsedAuthorizationHeader
}

type requestRecordTls struct {
	Version     uint16
	CipherSuite uint16
}

type requestRecordGeneve struct {
	VpcEndpointId string
	AttachmentId  string
	FlowCookie    string
}

type WebFirehose struct {
	enricher *lambda_acceptor.Enricher
	firehose firehoseiface.FirehoseAPI
	stream   string
	entries  chan []byte
}

func New(firehose firehoseiface.FirehoseAPI, ec2 ec2iface.EC2API, stream string) *WebFirehose {
	enricher, err := lambda_acceptor.NewEnricher(ec2)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	return &WebFirehose{
		firehose: firehose,
		stream:   stream,
		enricher: enricher,
		entries:  make(chan []byte),
	}
}

func (wf *WebFirehose) Run(ctx context.Context) error {
	batch := []*firehose.Record{}
	size := 0
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}

			// TODO: check batch count, combined batch size, etc
			_, err := wf.firehose.PutRecordBatchWithContext(ctx, &firehose.PutRecordBatchInput{
				DeliveryStreamName: &wf.stream,
				Records:            batch,
			})
			if err != nil {
				return errors.WithStack(err)
			}

			fmt.Printf("firehose dump: count=%d bytes=%d\n", len(batch), size)
			size = 0
			batch = nil
		case entry := <-wf.entries:
			fmt.Printf(string(entry))
			record := &firehose.Record{Data: entry}
			batch = append(batch, record)
			size += len(record.Data)
		}
	}
}

func (wf *WebFirehose) OnRequest(req *http.Request) {
}

func (wf *WebFirehose) OnResponse(resp *http.Response) error {
	ctx := resp.Request.Context()
	geneve := gwlb.GeneveOptionsFromContext(ctx)

	source, dest := gwlb.AddrsFromContext(ctx)
	var client, server *net.TCPAddr

	if source.Port == 443 || source.Port == 80 {
		server = source
		client = dest
	} else {
		server = dest
		client = source
	}

	var tlsrr *requestRecordTls
	if server.Port == 443 {
		tlsrr = &requestRecordTls{
			Version:     resp.Request.TLS.Version,
			CipherSuite: resp.Request.TLS.CipherSuite,
		}
	}

	var awsrr *requestRecordAws
	if parsed, ok := account_id_emf.ParseAuthorizationHeader(resp.Request.Header.Get("Authorization")); ok {
		awsrr = &requestRecordAws{
			AuthorizationHeader: parsed,
		}
	}

	host := resp.Request.URL.Host
	etldp1, _ := publicsuffix.EffectiveTLDPlusOne(host)

	rr := requestRecord{
		TLS: tlsrr,
		Geneve: requestRecordGeneve{
			VpcEndpointId: fmt.Sprintf("vpce-0%016x", geneve.VpcEndpointId),
			AttachmentId:  fmt.Sprintf("%016x", geneve.AttachmentId),
			FlowCookie:    fmt.Sprintf("%08x", geneve.FlowCookie),
		},
		AWS:            awsrr,
		Time:           time.Now(),
		Method:         resp.Request.Method,
		Host:           host,
		ETldPlusOne:    etldp1,
		URL:            resp.Request.URL.String(),
		RequestHeaders: flattenHeaders(resp.Request.Header),
		Client: lambda_acceptor.Endpoint{
			IpAddress: client.IP.String(),
			Port:      client.Port,
			EC2:       wf.ec2Endpoint(client.IP.String()),
		},

		StatusCode:      resp.StatusCode,
		StatusMessage:   resp.Status,
		ResponseHeaders: flattenHeaders(resp.Header),
		Server: lambda_acceptor.Endpoint{
			IpAddress: server.IP.String(),
			Port:      server.Port,
			EC2:       wf.ec2Endpoint(server.IP.String()),
		},
	}

	rrj, _ := json.Marshal(rr)
	wf.entries <- append(rrj, '\n')

	return nil
}

// TODO: this enriched info should be part of the ctx
// earlier on so it can be used by interceptors etc
func (wf *WebFirehose) ec2Endpoint(ip string) *lambda_acceptor.Ec2Endpoint {
	instance := wf.enricher.InstanceByIp(ip)
	if instance == nil {
		return nil
	}

	iface := instance.NetworkInterfaces[0]
	tags := map[string]string{}
	for _, tag := range instance.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return &lambda_acceptor.Ec2Endpoint{
		InstanceId: *instance.InstanceId,
		AccountId:  *iface.OwnerId,
		VpcId:      *iface.VpcId,
		SubnetId:   *iface.SubnetId,
		Tags:       tags,
	}
}

func flattenHeaders(head http.Header) map[string]string {
	flat := map[string]string{}

	for key, vals := range head {
		flat[key] = vals[0]
	}

	return flat
}
