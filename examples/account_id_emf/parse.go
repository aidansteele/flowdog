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
	// Credential is keyId/date/region/service/aws4_request; bail out on
	// anything malformed rather than panicking on a bad index.
	if len(s) < 4 {
		return ParsedAuthorizationHeader{}, false
	}

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

	// The algorithm indexes accessKeyId[4:13], so anything shorter is not a
	// valid AWS access key id and has no embedded account id.
	if len(accessKeyId) < 13 {
		return ""
	}

	// The account id is only embedded in AKIA/ASIA keys created after ~March
	// 2019, identifiable by accessKeyId[4] having a base32 value >= Q (i.e. in
	// the range Q-Z2-7). Older keys carry no internal structure - their id must
	// be looked up via sts:GetAccessKeyInfo - and running the algorithm on them
	// produces a bogus, often negative, account number.
	// See the "Conclusion, or so I thought" section of the linked post.
	if !strings.HasPrefix(accessKeyId, "AKIA") && !strings.HasPrefix(accessKeyId, "ASIA") {
		return ""
	}
	if strings.IndexByte(base32AwsFlavour, accessKeyId[4]) < strings.IndexByte(base32AwsFlavour, 'Q') {
		return ""
	}

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
