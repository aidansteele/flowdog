module github.com/aidansteele/flowdog

go 1.15

require (
	github.com/aws/aws-lambda-go v1.28.0
	github.com/aws/aws-sdk-go v1.42.31
	github.com/davecgh/go-spew v1.1.1
	github.com/disintegration/imaging v1.6.2
	github.com/glassechidna/go-emf v0.0.0-20220102031255-2c11928b55f0
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gopacket v1.1.19
	github.com/google/pprof v0.0.0-20220128192902-513e8ac6eea1 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/kenshaw/baseconv v0.1.1
	github.com/kor44/extcap v0.0.0-20201215145008-71f6cf07bb46
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/spinlock v0.1.0
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/genproto v0.0.0-20220112215332-a9c7c0acf9f2 // indirect
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.36.0
	inet.af/netstack v0.0.0-20211120045802-8aa80cf23d3c
	rogchap.com/v8go v0.7.0
)

replace (
	github.com/google/gopacket v1.1.19 => ./forks/gopacket
	rogchap.com/v8go v0.7.0 => github.com/rogchap/v8go v0.7.1-0.20220106173329-ede7cee433be
)
