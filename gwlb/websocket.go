package gwlb

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
)

func proxyWebsocket(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{CheckOrigin: nil} // TODO: check this field

	ctx := r.Context()
	port := ctx.Value(http.LocalAddrContextKey).(*net.TCPAddr).Port
	if port == 443 {
		r.URL.Scheme = "wss"
	} else if port == 80 {
		r.URL.Scheme = "ws"
	}

	newHeaders := r.Header.Clone()
	newHeaders.Del("Upgrade")
	newHeaders.Del("Connection")
	newHeaders.Del("Sec-Websocket-Key")
	newHeaders.Del("Sec-Websocket-Version")
	newHeaders.Del("Sec-Websocket-Extensions")
	newHeaders.Del("Sec-Websocket-Protocol")

	dialer := &websocket.Dialer{NetDialContext: DialContext}
	serverConn, resp, err := dialer.DialContext(ctx, r.URL.String(), newHeaders)
	if err != nil {
		fmt.Printf("wss dial %+v\n", err)
		panic(err)
	}

	clientConn, err := up.Upgrade(w, r, resp.Header)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return wsPumpLoop(ctx, serverConn, clientConn)
	})

	g.Go(func() error {
		return wsPumpLoop(ctx, clientConn, serverConn)
	})

	err = g.Wait()
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
}

func wsPumpLoop(ctx context.Context, a, b *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			mt, p, err := a.ReadMessage()
			if err != nil {
				return errors.WithStack(err)
			}

			err = b.WriteMessage(mt, p)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
}
