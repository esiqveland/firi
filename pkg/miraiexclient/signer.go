package miraiexclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

func NewSigner(clientId string, apiKey string, secret []byte) *signer {
	return &signer{
		clientId:       clientId,
		apiKey:         apiKey,
		secretKey:      secret,
		validForMillis: 2000,
	}
}

type SignedData struct {
	ClientID       string
	Signature      string
	Timestamp      time.Time
	ValidForMillis int64
}
type signer struct {
	apiKey         string
	clientId       string
	validForMillis int64
	secretKey      []byte
}

func (s *signer) Sign(ts time.Time) (*SignedData, error) {
	validForMillis := s.validForMillis

	type body struct {
		Timestamp      string `json:"timestamp"`
		ValidForMillis int64  `json:"validity"`
	}
	data, err := json.Marshal(&body{
		Timestamp: strconv.FormatInt(ts.Unix(), 10),
		//Timestamp:      ts.Unix(),
		ValidForMillis: validForMillis,
	})
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, s.secretKey)
	_, err = h.Write(data)
	if err != nil {
		return nil, err
	}

	sig := hex.EncodeToString(h.Sum(nil))
	signed := &SignedData{
		ClientID:       s.clientId,
		Signature:      sig,
		Timestamp:      ts,
		ValidForMillis: validForMillis,
	}
	return signed, nil
}
