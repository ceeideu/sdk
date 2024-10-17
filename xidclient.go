package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/ceeideu/sdk/crypto"
	"github.com/ceeideu/sdk/hem"
	"github.com/ceeideu/sdk/xid"
)

const (
	Xid         = "/xid"
	Generate    = "/generate"
	Map         = "/map"
	Lookup      = "/lookup"
	Decode      = "/decode"
	XidGenerate = Xid + Generate

	Keys        = "/keys"
	Token       = "/token"
	Refresh     = "/refresh"
	KeysRefresh = Keys + Refresh
	XidRefresh  = Xid + Refresh

	XApiKeyHeader = "x-api-key"
	XApiMockValue = "[X-API-VALUE]"

	SDKVersion = "x-sdk-version"
	SDKPrefix  = "go-sdk-"
)

type XID struct {
	baseURL       *url.URL
	authToken     string
	httpClient    HTTPDoer
	cryptoService Crypto
	SDKVersion    string
}

type Crypto interface {
	Encrypt(data []byte) (crypto.EncResp, error)
	Decrypt(keyID uint8, data []byte) ([]byte, error)
	KeysRefresh(keys crypto.Keys) error
}

var (
	ErrParse         = errors.New("parse error")
	ErrMarshal       = errors.New("marshal error")
	ErrRequest       = errors.New("request error")
	ErrCommunication = errors.New("communication error")
	ErrDecode        = errors.New("client decode error")
	ErrStatusNotOK   = errors.New("http status code not OK error")
	ErrDoHTTPReq     = errors.New("do http req error")
	ErrDecrypt       = errors.New("decrypt error")
	ErrOpen          = errors.New("open error")
	ErrConsent       = errors.New("no consent")
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithHTTPClient(c HTTPDoer) func(*XID) {
	return func(x *XID) {
		x.httpClient = c
	}
}

func NewXID(baseURL, authToken string, opts ...func(*XID)) (*XID, error) {
	sdkVer := "unknown"
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, d := range bi.Deps {
			if d.Path == "github.com/ceeideu/sdk" {
				sdkVer = d.Version
			}
		}
	}

	_xid := &XID{
		cryptoService: crypto.NewService(),
		SDKVersion:    SDKPrefix + sdkVer,
	}

	for _, o := range opts {
		o(_xid)
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
	}

	_xid.baseURL = u
	_xid.authToken = authToken

	if _xid.httpClient == nil {
		_xid.httpClient = http.DefaultClient
	}

	return _xid, nil
}

func (x *XID) DecryptToken(token xid.Token) (string, error) {
	keyID, err := token.Key()
	if err != nil {
		return "", fmt.Errorf("%s: %w", "key error", err)
	}

	_xid, err := token.XID()
	if err != nil {
		return "", fmt.Errorf("%s: %w", "xid error", err)
	}

	decrypted, err := x.cryptoService.Decrypt(keyID, _xid)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "decrypt error", err)
	}

	return string(decrypted), nil
}

func (x *XID) RefreshXID(ctx context.Context, refreshReq xid.RefreshReq) (xid.RefreshResp, error) {
	_bytes, err := json.Marshal(refreshReq)
	if err != nil {
		return xid.RefreshResp{}, fmt.Errorf("%w: %w", ErrMarshal, err)
	}

	resp, err := x.DoHTTPReq(ctx, http.MethodPost, x.baseURL.String()+XidRefresh, _bytes)
	if err != nil {
		return xid.RefreshResp{}, fmt.Errorf("%w:%w", ErrDoHTTPReq, err)
	}

	defer resp.Body.Close()

	var refreshResp xid.RefreshResp
	err = json.NewDecoder(resp.Body).Decode(&refreshResp)
	if err != nil {
		return xid.RefreshResp{}, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	return refreshResp, nil
}

func (x *XID) DoHTTPReq(ctx context.Context, method, _url string, body []byte) (*http.Response, error) {
	var bReader io.Reader
	if body != nil {
		bReader = bytes.NewBuffer(body)
	} else {
		bReader = http.NoBody
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		_url,
		bReader,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	req.Header.Set(XApiKeyHeader, x.authToken)
	req.Header.Set(SDKVersion, x.SDKVersion)

	resp, err := x.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCommunication, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w, got: %v", ErrStatusNotOK, resp.StatusCode)
	}

	return resp, nil
}

func (x *XID) Send(ctx context.Context, hemReq hem.Request) (xid.Response, error) {
	if hemReq.Err != nil {
		return xid.Response{}, fmt.Errorf("%s: %w", "hem request error", hemReq.Err)
	}

	_bytes, err := json.Marshal(hemReq)
	if err != nil {
		return xid.Response{}, fmt.Errorf("%w: %w", ErrMarshal, err)
	}

	resp, err := x.DoHTTPReq(ctx, http.MethodPost, x.baseURL.String()+XidGenerate, _bytes)
	if err != nil {
		return xid.Response{}, fmt.Errorf("%w: %w", ErrDoHTTPReq, err)
	}
	defer resp.Body.Close()

	xidResp := xid.Response{}

	err = json.NewDecoder(resp.Body).Decode(&xidResp)
	if err != nil {
		return xid.Response{}, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	return xidResp, nil
}

func (x *XID) TokenFromXID(_xid string) (xid.Token, error) {
	enc, err := x.cryptoService.Encrypt([]byte(_xid))
	if err != nil {
		return "", fmt.Errorf("%s: %w", "encrypt error", err)
	}

	return xid.NewToken(enc.EncKeyID, xid.Value(enc.Value)), nil
}

type KeysResp struct {
	Decryption map[uint8]string `json:"decryption"`
	Encryption Encryption       `json:"encryption"`
}

type Encryption struct {
	ID    uint8  `json:"id"`
	Value string `json:"value"`
}

func (x *XID) Refresh(ctx context.Context) error {
	resp, err := x.GetKeys(ctx)
	if err != nil {
		return err
	}

	err = x.cryptoService.KeysRefresh(crypto.Keys{
		Decryption: resp.Decryption,
		Encryption: crypto.Encryption{
			ID:    resp.Encryption.ID,
			Value: resp.Encryption.Value,
		},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "keys refresh error", err)
	}

	return nil
}

func (x *XID) GetKeys(ctx context.Context) (KeysResp, error) {
	resp, err := x.DoHTTPReq(ctx, http.MethodGet, x.baseURL.String()+KeysRefresh, nil)
	if err != nil {
		return KeysResp{}, fmt.Errorf("%w: %w", ErrDoHTTPReq, err)
	}
	defer resp.Body.Close()

	keyResp := KeysResp{}

	err = json.NewDecoder(resp.Body).Decode(&keyResp)
	if err != nil {
		return KeysResp{}, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	return keyResp, nil
}
