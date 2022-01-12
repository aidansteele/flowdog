package sts_rickroll

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

type StsRickroll struct{}

var stsRickrollRegexp = regexp.MustCompile(`<UserId>([^<]+)</UserId>`)

func (s *StsRickroll) OnRequest(req *http.Request) {}

func (s *StsRickroll) OnResponse(resp *http.Response) error {
	if resp.Request.Host != "sts.amazonaws.com" {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	body = stsRickrollRegexp.ReplaceAll(body, []byte(`<UserId>Never gonna give you up, never gonna let you down.</UserId>`))

	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))

	return nil
}
