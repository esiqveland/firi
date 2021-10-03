package miraiexclient

import (
	"testing"
	"time"
)

var secretKey = []byte("RTk2eNs67Vpan3345pmrwYEBYsWXRXtGF3BKTFq8WMLLOLOL")
var timestamp = time.Unix(1632954488, 0)

// See: https://developers.firi.com/#/authentication?id=hmac-encrypted-api-key
func TestSigner(t *testing.T) {
	expected := "d1845ce29857693063e7c1875dc513416b75747d617dbfe9a06f2b8ec87a4bc0"

	signer := NewSigner("", "", secretKey)

	s, err := signer.Sign(timestamp)
	if err != nil {
		t.Errorf("error signing: %v", err)
	}

	if s.Signature != expected {
		t.Errorf("error bad signature=%v expected=%v", s.Signature, expected)
	}
}
