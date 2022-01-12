package account_id_emf

import (
	"fmt"
	"github.com/glassechidna/go-emf/emf"
	"github.com/glassechidna/go-emf/emf/unit"
	"github.com/kenshaw/baseconv"
	"net/http"
	"strconv"
	"strings"
)

type AccountIdEmf struct{}

func (a *AccountIdEmf) OnRequest(req *http.Request) {
	prefix := "AWS4-HMAC-SHA256 Credential="

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return
	}

	s := strings.Split(strings.TrimPrefix(auth, prefix), "/")
	keyId, date, region, service := s[0], s[1], s[2], s[3]
	accountId := accountIdFromAccessKeyId(keyId)

	emf.Namespace = "flowdog"
	emf.Emit(emf.MSI{
		"KeyId":     keyId,
		"Date":      date,
		"Region":    region,
		"Service":   service,
		"AccountId": emf.Dimension(accountId),
		"Requests":  emf.Metric(1, unit.Count),
	})
}

func (a *AccountIdEmf) OnResponse(resp *http.Response) error {
	return nil
}

// from https://awsteele.com/blog/2020/09/26/aws-access-key-format.html
func accountIdFromAccessKeyId(accessKeyId string) string {
	base10 := "0123456789"
	base32AwsFlavour := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	offsetStr, _ := baseconv.Convert("QAAAAAAA", base32AwsFlavour, base10)
	offset, _ := strconv.Atoi(offsetStr)

	offsetAccountIdStr, _ := baseconv.Convert(accessKeyId[4:12], base32AwsFlavour, base10)
	offsetAccountId, _ := strconv.Atoi(offsetAccountIdStr)

	accountId := 2 * (offsetAccountId - offset)

	if strings.Index(base32AwsFlavour, accessKeyId[12:13]) >= strings.Index(base32AwsFlavour, "Q") {
		accountId++
	}

	return fmt.Sprintf("%012d", accountId)
}
