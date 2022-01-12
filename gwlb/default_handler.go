package gwlb

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

func DefaultHandler(interceptor Interceptor) http.Handler {
	rp := &httputil.ReverseProxy{
		ModifyResponse: interceptor.OnResponse,
		Director:       interceptor.OnRequest,
		Transport: &http.Transport{
			DialContext: DialContext,
			// these are copied from stdlib default
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		r.URL.Host = r.Host
		port := ctx.Value(http.LocalAddrContextKey).(*net.TCPAddr).Port
		if port == 443 {
			r.URL.Scheme = "https"
		} else if port == 80 {
			r.URL.Scheme = "http"
		}

		if r.Header.Get("Upgrade") == "websocket" {
			proxyWebsocket(w, r)
		} else {
			rp.ServeHTTP(w, r)
		}
	})

	return h
}
