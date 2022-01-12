package cloudfront_functions

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCloudfrontFunctions(t *testing.T) {
	cff, err := NewCloudfrontFunctions(`
		function onRequest(event) {
			return JSON.stringify(event.request);
		}
	`)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "https://google.com/abc?def=ghi", nil)
	req.Header.Set("Authorization", "monkey")
	cff.OnRequest(req)
}
