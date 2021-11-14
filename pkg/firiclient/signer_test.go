package firiclient

import (
	"testing"
	"time"
)

var secretKey = []byte("RTk2eNs67Vpan3345pmrwYEBYsWXRXtGF3BKTFq8WMLLOLOL")
var timestamp = time.Unix(1632954488, 0)

// See: https://developers.firi.com/#/authentication?id=hmac-encrypted-api-key
func TestSigner(t *testing.T) {
	expected := "21818f40878aad8ed30373d01ea7c5068f53ad47861a84dd3e6891f074ab681b"

	signer := NewSigner("", "", secretKey)

	s, err := signer.Sign(timestamp)
	if err != nil {
		t.Errorf("error signing: %v", err)
	}

	if s.Signature != expected {
		t.Errorf("error bad signature=%v expected=%v", s.Signature, expected)
	}
}
