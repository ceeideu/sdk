package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ceeideu/sdk/crypto"
	"github.com/ceeideu/sdk/hem"
	"github.com/ceeideu/sdk/xid"
)

var Err = errors.New("error")

func TestXID_Send(t *testing.T) {
	xidBytesMock := func() xid.Value {
		var bytes []byte
		bytes = append(bytes, 1, 0)
		bytes = append(bytes, []byte("foo")...)

		return bytes
	}()

	responseMock := func(resp xid.Response) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			m, _ := json.Marshal(resp)
			_, _ = w.Write(m)
		}
	}

	t.Parallel()
	tests := []struct {
		name        string
		hemReq      hem.Request
		handlerFunc http.HandlerFunc
		want        xid.Response
		decodeFn    func(string) ([]byte, error)
		wantErr     bool
	}{
		{
			name:   "ok",
			hemReq: hem.FromEmail("test@com"),
			handlerFunc: responseMock(
				xid.Response{
					Value: xidBytesMock.EncodeToString(),
				},
			),
			want: xid.Response{
				Value: xidBytesMock.EncodeToString(),
			},

			wantErr: false,
		},
		{
			name:        "err",
			hemReq:      hem.FromEmail("test@com"),
			handlerFunc: func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "error") },
			wantErr:     true,
		},
		{
			name:        "unauthorized",
			hemReq:      hem.FromEmail("test@com"),
			handlerFunc: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(test.handlerFunc)
			defer ts.Close()

			xidClient, err := NewXID(ts.URL, XApiInMemoryValue, WithHTTPClient(ts.Client()))
			if err != nil {
				t.Errorf("NewXID() error = %v", err)
			}
			got, err := xidClient.Send(context.Background(), test.hemReq)
			if (err != nil) != test.wantErr {
				t.Errorf("[%s] XID.Send() error = %v, wantErr %v", test.name, err, test.wantErr)

				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("XID.Send() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestXID_Refresh(t *testing.T) {
	t.Parallel()

	type fields struct {
		handlerFunc   http.HandlerFunc
		cryptoService Crypto
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "handler err",
			fields: fields{
				handlerFunc: func(writer http.ResponseWriter, r *http.Request) {
					writer.WriteHeader(http.StatusInternalServerError)
				},
				cryptoService: &CryptoMock{
					keysErr: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "crypto service error",
			fields: fields{
				handlerFunc: func(writer http.ResponseWriter, r *http.Request) {
					keysResp := KeysResp{
						Decryption: map[uint8]string{
							100: "foo",
							101: "bar",
						},
						Encryption: Encryption{
							ID:    101,
							Value: "bar",
						},
					}
					m, err := json.Marshal(keysResp)
					if err != nil {
						panic(err)
					}
					fmt.Fprint(writer, string(m))
				},
				cryptoService: &CryptoMock{
					keysErr: Err,
				},
			},
			wantErr: true,
		},
		{
			name: "with keys",
			fields: fields{
				handlerFunc: func(writer http.ResponseWriter, r *http.Request) {
					keysResp := KeysResp{
						Decryption: map[uint8]string{
							100: "foo",
							101: "bar",
						},
						Encryption: Encryption{
							ID:    101,
							Value: "bar",
						},
					}
					m, err := json.Marshal(keysResp)
					if err != nil {
						panic(err)
					}
					fmt.Fprint(writer, string(m))
				},
				cryptoService: &CryptoMock{
					keysErr: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		test := tt

		testServer := httptest.NewServer(test.fields.handlerFunc)
		xidClient, _ := NewXID(testServer.URL, "foo", WithHTTPClient(testServer.Client()))
		xidClient.cryptoService = tt.fields.cryptoService
		for i := 0; i < 1; i++ {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				if err := xidClient.Refresh(context.Background()); (err != nil) != test.wantErr {
					t.Errorf("Refresh() error = %v, wantErr %v", err, test.wantErr)
				}
			})
		}
	}
}

func TestXID_RefreshXId(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		handlerFunc http.HandlerFunc
		have        xid.RefreshReq
		want        xid.RefreshResp
		wantErr     bool
	}{
		{
			name: "valid flow",
			have: xid.RefreshReq{XID: "xid"},
			handlerFunc: func(writer http.ResponseWriter, req *http.Request) {
				type XIDValue struct {
					XID string `json:"xid"`
				}

				var xidValue XIDValue

				err := json.NewDecoder(req.Body).Decode(&xidValue)
				if err != nil {
					panic(err)
				}

				resp, err := json.Marshal(xid.RefreshResp{
					Value: xidValue.XID,
				})
				if err != nil {
					panic(err)
				}
				fmt.Fprint(writer, string(resp))
			},
			want: xid.RefreshResp{
				Value: "xid",
			},
			wantErr: false,
		},
		{
			name: "wrong format",
			have: xid.RefreshReq{XID: "xid"},
			handlerFunc: func(writer http.ResponseWriter, req *http.Request) {
				fmt.Fprint(writer, "bad format")
			},
			wantErr: true,
		},
		{
			name: "error from /xid/refresh endpoint",
			have: xid.RefreshReq{XID: "xid"},
			handlerFunc: func(writer http.ResponseWriter, req *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(test.handlerFunc)
			xidClient, _ := NewXID(testServer.URL, "foo", WithHTTPClient(testServer.Client()))

			got, err := xidClient.RefreshXID(context.Background(), test.have)
			if (err != nil) != test.wantErr {
				t.Errorf("[%s] XID.RegenerateXID() error = %v, wantErr %v", test.name, err, test.wantErr)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("[%s] XID.RegenerateXID() = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

type CryptoMock struct {
	encResp crypto.EncResp
	encErr  error
	decResp []byte
	decErr  error
	keysErr error
}

func (m *CryptoMock) Encrypt(_ []byte) (crypto.EncResp, error) {
	return m.encResp, m.encErr
}

func (m *CryptoMock) Decrypt(_ uint8, _ []byte) ([]byte, error) {
	return m.decResp, m.decErr
}

func (m *CryptoMock) KeysRefresh(_ crypto.Keys) error {
	return m.keysErr
}

func TestXID_DecryptToken(t *testing.T) {
	t.Parallel()
	type fields struct {
		crypto Crypto
	}
	type args struct {
		token xid.Token
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "key err",
			fields: fields{
				crypto: &CryptoMock{
					decErr: nil,
				},
			},
			args: args{
				token: xid.Token("not valid token"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "xid err",
			fields: fields{
				crypto: &CryptoMock{
					decErr: nil,
				},
			},
			args: args{
				token: xid.Token("2not valid token"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "decrypt err",
			fields: fields{
				crypto: &CryptoMock{
					decErr: Err,
				},
			},
			args: args{
				token: xid.NewToken(2, xid.Value("foo")),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "ok",
			fields: fields{
				crypto: &CryptoMock{
					decErr:  nil,
					decResp: []byte("ok"),
				},
			},
			args: args{
				token: xid.NewToken(2, xid.Value("foo")),
			},
			want:    "ok",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			x := &XID{
				cryptoService: test.fields.crypto,
			}
			got, err := x.DecryptToken(test.args.token)
			if (err != nil) != test.wantErr {
				t.Errorf("DecryptToken() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if got != test.want {
				t.Errorf("DecryptToken() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestXID_TokenFromXID(t *testing.T) {
	t.Parallel()
	type fields struct {
		cryptoService Crypto
	}
	type args struct {
		_xid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    xid.Token
		wantErr bool
	}{
		{
			name: "empty",
			fields: fields{
				cryptoService: &CryptoMock{
					encResp: crypto.EncResp{},
					encErr:  nil,
				},
			},
			args: args{
				_xid: "",
			},
			want:    "0",
			wantErr: false,
		},
		{
			name: "enc error",
			fields: fields{
				cryptoService: &CryptoMock{
					encResp: crypto.EncResp{},
					encErr:  Err,
				},
			},
			args: args{
				_xid: "some xid",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "ok",
			fields: fields{
				cryptoService: &CryptoMock{
					encResp: crypto.EncResp{
						Value:    []byte("some xid"),
						EncKeyID: 1,
					},
					encErr: nil,
				},
			},
			args: args{
				_xid: "some xid",
			},
			want:    xid.Token("1" + base64.StdEncoding.EncodeToString([]byte("some xid"))),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			x := &XID{
				cryptoService: test.fields.cryptoService,
			}
			got, err := x.TokenFromXID(test.args._xid)
			if (err != nil) != test.wantErr {
				t.Errorf("XID.TokenFromXID() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("XID.TokenFromXID() = %v, want %v", got, test.want)
			}
		})
	}
}
