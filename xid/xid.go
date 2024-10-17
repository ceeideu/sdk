package xid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"

	"github.com/ceeideu/sdk/properties"
)

func Hash(b []byte) []byte {
	h := sha256.New()
	h.Write(b)

	return h.Sum(nil)
}

const (
	// unique value size.
	LenValue = 8
	// metadata size.
	LenMeta = 2
	// xid size.
	Len = LenMeta + LenValue

	IdxVersion = 0
	IdxTypeOf  = 1
	// value begins where metadata ends.
	IdxValue = LenMeta

	TypeOfUnknown = TypeOf(0)
	TypeOfEmail   = TypeOf(1)
	TypeOfHex     = TypeOf(2)
	TypeOfLTID    = TypeOf(3)

	Email = "email"
	Phone = "phone"
	Hex   = "hex"
	LTID  = "ltid"

	StatusOfUnknown        = StatusOf(0)
	StatusOfOK             = StatusOf(1)
	StatusOfUserBlocked    = StatusOf(2)
	StatusOfInvalidConsent = StatusOf(3)

	Okay           = "ok"
	UserBlocked    = "userBlocked"
	InvalidConsent = "invalid consent"

	Unknown = "unknown"
)

type TypeOf byte

func TypeFromString(s string) TypeOf {
	switch s {
	case Email:
		return TypeOfEmail
	case Hex:
		return TypeOfHex
	case LTID:
		return TypeOfLTID
	default:
		return TypeOfUnknown
	}
}

func (t TypeOf) String() string {
	switch t {
	case TypeOfEmail:
		return Email
	case TypeOfHex:
		return Hex
	case TypeOfLTID:
		return LTID
	case TypeOfUnknown:
		return Unknown
	default:
		return Unknown
	}
}

type StatusOf byte

func (t StatusOf) String() string {
	switch t {
	case StatusOfOK:
		return Okay
	case StatusOfUserBlocked:
		return UserBlocked
	case StatusOfUnknown:
		return Unknown
	case StatusOfInvalidConsent:
		return InvalidConsent
	default:
		return Unknown
	}
}

type Value []byte

// Constructor function.
type CtorFunc func(version byte, typeof TypeOf) (Value, error)

func randRead(buf []byte) error {
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("%s: %w", "read error", err)
	}

	return nil
}

func Rand(version byte, t TypeOf) (Value, error) {
	return val(version, t, randRead)
}

func val(version byte, typeof TypeOf, fn func([]byte) error) (Value, error) {
	buf := make([]byte, Len)

	if err := fn(buf[LenMeta:]); err != nil {
		return nil, err
	}

	buf[IdxVersion] = version
	buf[IdxTypeOf] = byte(typeof)

	return buf, nil
}

func (x Value) Version() byte {
	if len(x) != Len {
		return byte(0)
	}

	return x[IdxVersion]
}

func (x Value) Type() byte {
	if len(x) != Len {
		return byte(TypeOfUnknown)
	}

	return x[IdxTypeOf]
}

func (x Value) EncodeToString() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(x)
}

func Decode(s string /* xid0 */) (Value, error) {
	v, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)

	return Value(v), err
}

type RefreshReq struct {
	XID        string            `json:"xid"`
	Properties map[string]string `json:"properties"`
}

func RefreshRequest(xid string) RefreshReq {
	return RefreshReq{XID: xid}
}

func (r RefreshReq) WithProperties(_properties properties.Value) RefreshReq {
	r.Properties = _properties

	return r
}

type RefreshResp struct {
	Value string `json:"value"`
}

type Response struct {
	Value  string `json:"value"`
	Status string `json:"status"`
}
