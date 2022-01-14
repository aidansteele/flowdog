package cloudfront_functions

import (
	"github.com/aidansteele/flowdog/gwlb"
	"io/ioutil"
)

func NewRickroll() gwlb.Interceptor {
	script, _ := ioutil.ReadFile("rick.js")
	cff, _ := NewCloudfrontFunctions(string(script))
	return cff
}
