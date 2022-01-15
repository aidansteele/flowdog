package main

import (
	"context"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb/shark"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/kor44/extcap"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"os"
	"strings"
)

var sslKeyLogFileOpt *extcap.ConfigStringOpt

func main() {
	// TODO: open upstream github issue about capture filter validation not working
	// note that program is invoked for every keystroke in capture filter textbox
	if len(os.Args) > 1 && os.Args[1] == "--extcap-capture-filter" {
		_, err := shark.FilterVM(os.Args[2])
		if err == nil {
			os.Exit(0)
		}
	}

	// TODO: open upstream issue about this unrecognised flag warning
	versionIdx := -1
	for idx, arg := range os.Args {
		if strings.HasPrefix(arg, "--extcap-version") {
			versionIdx = idx
		}
	}
	if versionIdx > 0 {
		os.Args = append(os.Args[:versionIdx], os.Args[versionIdx+1:]...)
	}

	sslKeyLogFileOpt = extcap.NewConfigStringOpt("sslkeylogfile", "confdisplay")
	sslKeyLogFileOpt.Default("/tmp/sslkeylogfile")
	sslKeyLogFileOpt.Placeholder("the placeholder")
	sslKeyLogFileOpt.Tooltip("the tooltip")

	app := extcap.App{
		Usage:         "flowdogshark",
		HelpPage:      "flowdogshark attaches to flowdog-managed AWS GWLB appliances for VPC-wide packet capture",
		GetInterfaces: getAllInterfaces,
		GetDLT:        getDLT,
		StartCapture:  startCapture,
		GetAllConfigOptions: func() []extcap.ConfigOption {
			return []extcap.ConfigOption{sslKeyLogFileOpt}
		},
		GetConfigOptions: func(iface string) ([]extcap.ConfigOption, error) {
			return []extcap.ConfigOption{sslKeyLogFileOpt}, nil
		},
	}

	app.Run(os.Args)
}

func getAllInterfaces() ([]extcap.CaptureInterface, error) {
	return []extcap.CaptureInterface{
		{
			Value:   "vpce-08adcae6a2example",
			Display: "flowdogshark: display",
		},
	}, nil
}

func getDLT(iface string) (extcap.DLT, error) {
	return extcap.DLT{
		Number:  int(layers.LinkTypeRaw),
		Name:    "LINKTYPE_RAW",
		Display: "dlt display?",
	}, nil
}

type stringifier interface {
	String() string
}

func startCapture(iface string, pipe io.WriteCloser, filter string, opts map[string]interface{}) error {
	var keyLogFile *os.File
	// TODO: upstream issue about how tricky it is to extract opts
	if keyLogPath, ok := opts["sslkeylogfile"].(stringifier); ok && keyLogPath.String() != "" {
		var err error
		keyLogFile, err = os.OpenFile(keyLogPath.String(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return errors.WithStack(err)
		}
		defer keyLogFile.Close()
	}

	defer pipe.Close()
	w := pcapgo.NewWriter(pipe)

	err := w.WriteFileHeader(0, layers.LinkTypeRaw)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:7081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	client := shark.NewVpcsharkClient(conn)
	cc, err := client.GetPackets(ctx, &shark.GetPacketsInput{Filter: filter})
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	for {
		msg, err := cc.Recv()
		if err != nil {
			return errors.WithStack(err)
		}

		if keyLogFile != nil && msg.SslKeyLog != nil {
			_, err = keyLogFile.Write(msg.SslKeyLog)
			if err != nil {
				return errors.WithStack(err)
			}

			keyLogFile.Sync()
		}

		if msg.Payload != nil {
			gpkt := gopacket.NewPacket(msg.Payload, layers.LayerTypeGeneve, gopacket.Default)
			geneve := gpkt.Layer(layers.LayerTypeGeneve).(*layers.Geneve)
			payload := geneve.LayerPayload()

			err = w.WritePacket(gopacket.CaptureInfo{
				Timestamp:      msg.Time.AsTime(),
				CaptureLength:  len(payload),
				Length:         len(payload),
				InterfaceIndex: 0,
			}, payload)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
