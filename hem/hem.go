package hem

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/ceeideu/sdk/properties"
	"github.com/ceeideu/sdk/xid"
)

var (
	ErrParseEmail = errors.New("parse email error")
	ErrLen        = errors.New("len error")
)

type Request struct {
	Type       string            `json:"type"`
	Value      string            `json:"value"`
	Err        error             `json:"-"`
	Properties map[string]string `json:"properties"`
}

func (r Request) WithProperties(_properties properties.Value) Request {
	r.Properties = _properties

	return r
}

func FromEmail(email string) Request {
	em, err := NormalizeEmail(email)

	return Request{
		Type:  xid.Email,
		Value: base64.StdEncoding.EncodeToString(xid.Hash([]byte(em))),
		Err:   err,
	}
}

func FromHex(hem string) Request {
	const (
		// base16 encoded length of 32 bytes
		hexLen = 64
	)

	if len(hem) != hexLen {
		return Request{Err: ErrLen}
	}

	if _, err := hex.DecodeString(hem); err != nil {
		return Request{Err: fmt.Errorf("%s: %w", "parse error", err)}
	}

	return Request{Type: xid.Hex, Value: hem}
}

func NormalizeEmail(email string) (string, error) {
	m, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("%w %w", ErrParseEmail, err)
	}

	email = strings.ToLower(m.Address)

	emailParts := strings.Split(email, "@")
	normLocal := emailParts[0]
	normDomain := emailParts[1]

	p := strings.Index(normLocal, "+")
	if p != -1 {
		normLocal = normLocal[:p]
	}

	if normDomain == "gmail.com" {
		normLocal = strings.ReplaceAll(normLocal, ".", "")
	}

	return normLocal + "@" + normDomain, nil
}
