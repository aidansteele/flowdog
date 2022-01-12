package geneve_headers

import (
	"fmt"
	"github.com/aidansteele/flowdog/gwlb"
	"net/http"
)

type GeneveHeaders struct{}

func (g *GeneveHeaders) OnRequest(req *http.Request) {
	opts := gwlb.GeneveOptionsFromContext(req.Context())
	addHeaders(req.Header, opts)
}

func (g *GeneveHeaders) OnResponse(resp *http.Response) error {
	opts := gwlb.GeneveOptionsFromContext(resp.Request.Context())
	addHeaders(resp.Header, opts)
	return nil
}

func addHeaders(header http.Header, opts gwlb.AwsGeneveOptions) {
	header.Set("X-Gwlb-Vpc-Endpoint-Id", fmt.Sprintf("vpce-0%016x", opts.VpcEndpointId))
	header.Set("X-Gwlb-Attachment-Id", fmt.Sprintf("%016x", opts.AttachmentId))
	header.Set("X-Gwlb-Flow-Cookie", fmt.Sprintf("%08x", opts.FlowCookie))
}
