package account_id_emf

import (
	"fmt"
	"github.com/kenshaw/baseconv"
	"strconv"
	"strings"
)

type ParsedAuthorizationHeader struct {
	KeyId     string
	Date      string
	Region    string
	Service   string
	AccountId string
}

func ParseAuthorizationHeader(auth string) (ParsedAuthorizationHeader, bool) {
	prefix := "AWS4-HMAC-SHA256 Credential="
	if !strings.HasPrefix(auth, prefix) {
		return ParsedAuthorizationHeader{}, false
	}

	s := strings.Split(strings.TrimPrefix(auth, prefix), "/")
	keyId, date, region, service := s[0], s[1], s[2], s[3]
	accountId := accountIdFromAccessKeyId(keyId)
	return ParsedAuthorizationHeader{
		KeyId:     keyId,
		Date:      date,
		Region:    region,
		Service:   service,
		AccountId: accountId,
	}, true
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
