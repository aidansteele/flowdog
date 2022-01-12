module github.com/aidansteele/flowdog

go 1.15

require (
	github.com/aws/aws-lambda-go v1.28.0 // indirect
	github.com/aws/aws-sdk-go v1.42.31
	github.com/davecgh/go-spew v1.1.1
	github.com/glassechidna/go-emf v0.0.0-20220102031255-2c11928b55f0
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/kenshaw/baseconv v0.1.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/tidwall/spinlock v0.1.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	inet.af/netstack v0.0.0-20211120045802-8aa80cf23d3c
	rogchap.com/v8go v0.7.0
)

replace rogchap.com/v8go v0.7.0 => github.com/rogchap/v8go v0.7.1-0.20220106173329-ede7cee433be
