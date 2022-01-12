package gwlb

import (
	"context"
	"inet.af/netstack/tcpip/stack"
	"net"
)

type contextKey string

const (
	contextKeyGeneveOptions = contextKey("contextKeyGeneveOptions")
	contextKeySourceAddr    = contextKey("contextKeySourceAddr")
	contextKeyNetstack      = contextKey("contextKeyNetstack")
)

func ContextWithGeneveOptions(ctx context.Context, opts AwsGeneveOptions) context.Context {
	return context.WithValue(ctx, contextKeyGeneveOptions, opts)
}

func GeneveOptionsFromContext(ctx context.Context) AwsGeneveOptions {
	opts, _ := ctx.Value(contextKeyGeneveOptions).(AwsGeneveOptions)
	return opts
}

func ContextWithSourceAddr(ctx context.Context, addr *net.TCPAddr) context.Context {
	return context.WithValue(ctx, contextKeySourceAddr, addr)
}

func SourceAddrFromContext(ctx context.Context) *net.TCPAddr {
	return ctx.Value(contextKeySourceAddr).(*net.TCPAddr)
}

func ContextWithNetstack(ctx context.Context, netstack *stack.Stack) context.Context {
	return context.WithValue(ctx, contextKeyNetstack, netstack)
}

func NetstackFromContext(ctx context.Context) *stack.Stack {
	return ctx.Value(contextKeyNetstack).(*stack.Stack)
}
