package upsidedown

import (
	"bytes"
	"fmt"
	"github.com/aidansteele/flowdog/gwlb"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"strings"
)

type interceptor struct{}

// all credit to http://www.ex-parrot.com/pete/upside-down-ternet.html

func UpsideDown() gwlb.Interceptor {
	return &interceptor{}
}

func (i *interceptor) OnRequest(req *http.Request) {
	// content ranges make this prank more work
	req.Header.Del("Range")
	req.Header.Del("If-Range")
}

func (i *interceptor) OnResponse(resp *http.Response) error {
	resp.Header.Del("Accept-Ranges")

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil
	}

	originalBody := resp.Body
	defer originalBody.Close()

	img, err := imaging.Decode(originalBody)
	if err != nil {
		return errors.WithStack(err)
	}

	// perform the.. "magic"
	img = imaging.Rotate180(img)
	img = imaging.Blur(img, 1.5)

	newBody := &bytes.Buffer{}
	resp.Body = ioutil.NopCloser(newBody)

	// i'm lazy so lets make everything a png
	err = imaging.Encode(newBody, img, imaging.PNG)
	if err != nil {
		return errors.WithStack(err)
	}

	resp.Header.Set("Content-Type", "image/png")
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", newBody.Len()))
	return nil
}
