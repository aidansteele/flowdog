package cloudfront_functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"rogchap.com/v8go"
	"strings"
)

type CloudfrontFunctions struct {
	reqV8 *v8go.Context
	reqFn *v8go.Function
}

type cffContextKey string

const cffContextKeyNewResp = cffContextKey("cffContextKeyNewResp")

func NewCloudfrontFunctions(onRequest string) (*CloudfrontFunctions, error) {
	v8 := v8go.NewContext()
	_, err := v8.RunScript(onRequest, "onRequest.js")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fnv, err := v8.Global().Get("onRequest")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fn, err := fnv.AsFunction()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CloudfrontFunctions{reqV8: v8, reqFn: fn}, nil
}

func (c *CloudfrontFunctions) OnRequest(req *http.Request) {
	qs := map[string]cffValue{}
	for key, values := range req.URL.Query() {
		qs[key] = cffValue{Value: values[0]}
	}

	hdr := map[string]cffValue{}
	for key, values := range req.Header {
		hdr[strings.ToLower(key)] = cffValue{Value: values[0]}
	}
	hdr["host"] = cffValue{Value: req.URL.Host}

	ck := map[string]cffValue{}

	payload := cffEvent{
		Version: "1.0",
		Context: cffContext{EventType: "viewer-request"},
		Request: cffRequest{
			Method:      req.Method,
			Uri:         req.URL.Path,
			Querystring: qs,
			Headers:     hdr,
			Cookies:     ck,
		},
	}

	j, _ := json.Marshal(payload)
	val, err := v8go.JSONParse(c.reqV8, string(j))
	if err != nil {
		fmt.Printf("%+v\n", errors.WithStack(err))
		panic(err)
	}

	retval, err := c.reqFn.Call(c.reqV8.Global(), val)
	if err != nil {
		fmt.Printf("%+v\n", errors.WithStack(err))
		panic(err)
	}

	js, err := v8go.JSONStringify(c.reqV8, retval)
	if err != nil {
		fmt.Printf("%+v\n", errors.WithStack(err))
		panic(err)
	}

	newresp := cffResponse{}
	_ = json.Unmarshal([]byte(js), &newresp)

	if newresp.StatusCode > 0 {
		// TODO: this doesn't actually stop the original request from being fired.
		fmt.Println("editing response instead")
		ctx := context.WithValue(req.Context(), cffContextKeyNewResp, &newresp)
		reqctx := req.WithContext(ctx)
		*req = *reqctx
	} else {
		err = c.modifyRequest(req, js)
		if err != nil {
			fmt.Printf("%+v\n", err)
			panic(err)
		}
	}
}

func (c *CloudfrontFunctions) modifyRequest(req *http.Request, js string) error {
	newreq := cffRequest{}
	err := json.Unmarshal([]byte(js), &newreq)
	if err != nil {
		return errors.WithStack(err)
	}

	req.Method = newreq.Method
	req.Header = http.Header{}
	for key, value := range newreq.Headers {
		req.Header.Set(key, value.Value)
	}

	newhost := newreq.Headers["host"].Value
	req.URL.Host = newhost
	req.Host = req.URL.Host
	req.URL.Path = newreq.Uri

	q := url.Values{}
	for key, value := range newreq.Querystring {
		q.Set(key, value.Value)
	}
	req.URL.RawQuery = q.Encode()

	// TODO: cookies
	// TODO: body
	return nil
}

func (c *CloudfrontFunctions) OnResponse(resp *http.Response) error {
	newresp, ok := resp.Request.Context().Value(cffContextKeyNewResp).(*cffResponse)
	if !ok {
		return nil
	}

	fmt.Println("returning edited response")
	resp.StatusCode = newresp.StatusCode
	resp.Status = fmt.Sprintf("%d %s", newresp.StatusCode, newresp.StatusDescription)
	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

	resp.Header = http.Header{}
	for key, value := range newresp.Headers {
		resp.Header.Set(key, value.Value)
	}

	return nil
}

type cffContext struct {
	DistributionDomainName string `json:"distributionDomainName"`
	DistributionId         string `json:"distributionId"`
	EventType              string `json:"eventType"`
	RequestId              string `json:"requestId"`
}

type cffValue struct {
	Value string `json:"value"`
}

type cffRequest struct {
	Method      string              `json:"method"`
	Uri         string              `json:"uri"`
	Querystring map[string]cffValue `json:"querystring"`
	Headers     map[string]cffValue `json:"headers"`
	Cookies     map[string]cffValue `json:"cookies"`
}

type cffSetCookie struct {
	Value      string `json:"value"`
	Attributes string `json:"attributes"`
}

type cffResponse struct {
	StatusCode        int                     `json:"statusCode"`
	StatusDescription string                  `json:"statusDescription"`
	Headers           map[string]cffValue     `json:"headers"`
	Cookies           map[string]cffSetCookie `json:"cookies"`
}

type cffEvent struct {
	Version  string      `json:"version"`
	Context  cffContext  `json:"context"`
	Request  cffRequest  `json:"request,omitempty"`
	Response cffResponse `json:"response,omitempty"`
	Viewer   struct {
		Ip string `json:"ip"`
	} `json:"viewer"`
}
