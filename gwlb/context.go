package gwlb

import (
	"context"
	"inet.af/netstack/tcpip/stack"
	"net"
)

type contextKey string

const (
	contextKeyGeneveOptions = contextKey("contextKeyGeneveOptions")
	contextKeyAddrs         = contextKey("contextKeyAddrs")
	contextKeyNetstack      = contextKey("contextKeyNetstack")
)

func ContextWithGeneveOptions(ctx context.Context, opts AwsGeneveOptions) context.Context {
	return context.WithValue(ctx, contextKeyGeneveOptions, opts)
}

func GeneveOptionsFromContext(ctx context.Context) AwsGeneveOptions {
	opts, _ := ctx.Value(contextKeyGeneveOptions).(AwsGeneveOptions)
	return opts
}

func ContextWithAddrs(ctx context.Context, source, dest *net.TCPAddr) context.Context {
	slice := []*net.TCPAddr{source, dest}
	return context.WithValue(ctx, contextKeyAddrs, slice)
}

func AddrsFromContext(ctx context.Context) (source *net.TCPAddr, dest *net.TCPAddr) {
	slice := ctx.Value(contextKeyAddrs).([]*net.TCPAddr)
	return slice[0], slice[1]
}

func ContextWithNetstack(ctx context.Context, netstack *stack.Stack) context.Context {
	return context.WithValue(ctx, contextKeyNetstack, netstack)
}

func NetstackFromContext(ctx context.Context) *stack.Stack {
	return ctx.Value(contextKeyNetstack).(*stack.Stack)
}
