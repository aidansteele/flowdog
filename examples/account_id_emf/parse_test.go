package account_id_emf

import (
	"encoding/base32"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

const awsBase32 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// canonicalAccountId is an independent implementation of the algorithm
// described at https://awsteele.com/blog/2020/09/26/aws-access-key-format.html
// It base32-decodes the key id and extracts the embedded account id, so we can
// cross-check the (baseconv-based) implementation in parse.go.
func canonicalAccountId(t *testing.T, accessKeyId string) string {
	t.Helper()
	dec, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(accessKeyId[4:])
	if err != nil {
		t.Fatalf("base32 decode %q: %v", accessKeyId, err)
	}
	var z uint64
	for _, b := range dec[:6] {
		z = z<<8 | uint64(b)
	}
	accountId := (z & 0x7fffffffff80) >> 7
	return fmt.Sprintf("%012d", accountId)
}

// TestAccountIdFromAccessKeyId_KnownVector uses the real access key id /
// account id pair that sts:GetAccessKeyInfo returned in the source blog post.
func TestAccountIdFromAccessKeyId_KnownVector(t *testing.T) {
	const key = "ASIAY34FZKBOKMUTVV7A"
	const want = "609629065308"

	if got := accountIdFromAccessKeyId(key); got != want {
		t.Fatalf("accountIdFromAccessKeyId(%q) = %q, want %q", key, got, want)
	}
	// sanity-check our independent implementation against the same known pair
	if got := canonicalAccountId(t, key); got != want {
		t.Fatalf("canonicalAccountId(%q) = %q, want %q", key, got, want)
	}
}

func TestAccountIdFromAccessKeyId_MatchesCanonical(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	qIdx := strings.IndexByte(awsBase32, 'Q')

	for i := 0; i < 10000; i++ {
		suffix := make([]byte, 16)
		// keyid[4] (suffix[0]) must be >= Q for a modern account-bearing key
		suffix[0] = awsBase32[qIdx+rng.Intn(len(awsBase32)-qIdx)]
		for j := 1; j < len(suffix); j++ {
			suffix[j] = awsBase32[rng.Intn(len(awsBase32))]
		}
		key := "AKIA" + string(suffix)

		got := accountIdFromAccessKeyId(key)
		want := canonicalAccountId(t, key)
		if got != want {
			t.Fatalf("key %q: got %s want %s", key, got, want)
		}
	}
}

// TestAccountIdFromAccessKeyId_OldFormatKeys ensures pre-2019 keys (keyid[4] < Q)
// return empty rather than a bogus negative account id.
func TestAccountIdFromAccessKeyId_OldFormatKeys(t *testing.T) {
	for _, key := range []string{
		"AKIAJ34FZKBOKMUTVV7A", // J < Q
		"AKIAIOSFODNN7EXAMPLE", // classic docs example, I < Q
		"1YRA5YCR63BKA0BX35G2", // pre-2010, no AKIA/ASIA prefix
	} {
		if got := accountIdFromAccessKeyId(key); got != "" {
			t.Fatalf("accountIdFromAccessKeyId(%q) = %q, want empty (unsupported)", key, got)
		}
	}
}

func TestAccountIdFromAccessKeyId_ShortInputDoesNotPanic(t *testing.T) {
	for _, in := range []string{"", "AKIA", "AKIAABCD", "AKIAABCDEFG"} {
		if got := accountIdFromAccessKeyId(in); got != "" {
			t.Fatalf("accountIdFromAccessKeyId(%q) = %q, want empty", in, got)
		}
	}
}

func TestParseAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"not sigv4", "Bearer abc123", false},
		{"empty", "", false},
		{"missing fields", "AWS4-HMAC-SHA256 Credential=ASIAY34FZKBOKMUTVV7A/20130524", false},
		{
			"valid",
			"AWS4-HMAC-SHA256 Credential=ASIAY34FZKBOKMUTVV7A/20130524/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=abc",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, ok := ParseAuthorizationHeader(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok {
				if parsed.Region != "us-east-1" || parsed.Service != "s3" || parsed.Date != "20130524" {
					t.Fatalf("unexpected parse result: %+v", parsed)
				}
				if parsed.AccountId != "609629065308" {
					t.Fatalf("unexpected account id: %q", parsed.AccountId)
				}
			}
		})
	}
}
