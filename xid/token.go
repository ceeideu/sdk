package xid

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

var ErrTokenLen = errors.New("token len err")

type TokenResponse struct {
	Value string `json:"value"`
}

type TokenRefreshReq struct {
	XID string `json:"xid"`
}

type TokenRefreshResp struct {
	Token string `json:"token"`
}

type Token string

func NewToken(keyID uint8, xid Value) Token {
	tkn := strconv.Itoa(int(keyID)) + base64.StdEncoding.EncodeToString(xid)

	return Token(tkn)
}

func (t Token) XID() ([]byte, error) {
	token := t.String()
	if len(token) <= 1 {
		return nil, ErrTokenLen
	}

	v, err := base64.StdEncoding.DecodeString(token[1:])
	if err != nil {
		return v, fmt.Errorf("%s: %w", "decode err", err)
	}

	return v, nil
}

func (t Token) Key() (uint8, error) {
	key, err := strconv.ParseUint(string(t[0]), 10, 8)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "uint err", err)
	}

	return uint8(key), nil
}

func (t Token) String() string {
	return string(t)
}
