package gwlb

import "net/http"

// Interceptor is modelled on https://pkg.go.dev/net/http/httputil#ReverseProxy
type Interceptor interface {
	OnRequest(req *http.Request)
	OnResponse(resp *http.Response) error
}

type Chain []Interceptor

func (c Chain) OnRequest(req *http.Request) {
	for _, interceptor := range c {
		interceptor.OnRequest(req)
	}
}

func (c Chain) OnResponse(resp *http.Response) error {
	for _, interceptor := range c {
		err := interceptor.OnResponse(resp)
		if err != nil {
			return err
		}
	}

	return nil
}
