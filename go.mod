module github.com/aidansteele/flowdog

go 1.25.0

require (
	github.com/aws/aws-lambda-go v1.28.0
	github.com/aws/aws-sdk-go v1.42.31
	github.com/disintegration/imaging v1.6.2
	github.com/glassechidna/go-emf v0.0.0-20220102031255-2c11928b55f0
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.4.2
	github.com/kenshaw/baseconv v0.1.1
	github.com/kor44/extcap v0.0.0-20201215145008-71f6cf07bb46
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/spinlock v0.1.0
	golang.org/x/net v0.52.0
	golang.org/x/sync v0.20.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	gvisor.dev/gvisor v0.0.0-20240930222854-8b2593668a9f
	rogchap.com/v8go v0.7.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/urfave/cli/v2 v2.2.0 // indirect
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/google/gopacket v1.1.19 => ./forks/gopacket
	rogchap.com/v8go v0.7.0 => github.com/rogchap/v8go v0.7.1-0.20220106173329-ede7cee433be
)
