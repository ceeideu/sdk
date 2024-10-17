package crypto

import (
	"crypto/cipher"
	"encoding/hex"
	"io"
	"reflect"
	"sync"
	"testing"
)

const (
	hexString = "000000000000000000000000b99e66d5e913c47273857540395258aeefd048"
)

func TestService_KeysRefresh(t *testing.T) {
	t.Parallel()
	type fields struct {
		cipherKeys CipherKeys
	}
	type args struct {
		keys Keys
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantRepo CipherKeys
	}{
		{
			name: "empty repo",
			fields: fields{
				cipherKeys: CipherKeys{},
			},
			args: args{
				keys: Keys{
					Decryption: map[uint8]string{
						100: hex.EncodeToString([]byte("0000000000000foo")),
						101: hex.EncodeToString([]byte("0000000000000bar")),
					},
					Encryption: Encryption{
						ID:    101,
						Value: hex.EncodeToString([]byte("0000000000000bar")),
					},
				},
			},
			wantErr: false,
			wantRepo: CipherKeys{
				Decryption: map[uint8]cipher.AEAD{
					100: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000foo")))

						return key
					}(),
					101: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

						return key
					}(),
				},
				Encryption: Enc{
					ID: 101,
					Value: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

						return key
					}(),
				},
			},
		},
		{
			name: "new keys",
			fields: fields{
				cipherKeys: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						100: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000foo")))

							return key
						}(),
						101: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

							return key
						}(),
					},
					Encryption: Enc{
						ID: 101,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

							return key
						}(),
					},
				},
			},
			args: args{
				keys: Keys{
					Decryption: map[uint8]string{
						200: hex.EncodeToString([]byte("0000000000000bar")),
						201: hex.EncodeToString([]byte("0000000000000baz")),
					},
					Encryption: Encryption{
						ID:    201,
						Value: hex.EncodeToString([]byte("0000000000000baz")),
					},
				},
			},
			wantErr: false,
			wantRepo: CipherKeys{
				Decryption: map[uint8]cipher.AEAD{
					200: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

						return key
					}(),
					201: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000baz")))

						return key
					}(),
				},
				Encryption: Enc{
					ID: 201,
					Value: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000baz")))

						return key
					}(),
				},
			},
		},
		{
			name: "error",
			fields: fields{
				cipherKeys: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						100: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000foo")))

							return key
						}(),
						101: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

							return key
						}(),
					},
					Encryption: Enc{
						ID: 101,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

							return key
						}(),
					},
				},
			},
			args: args{
				keys: Keys{
					Decryption: map[uint8]string{
						200: hex.EncodeToString([]byte("badKey")),
						201: hex.EncodeToString([]byte("badKey")),
					},
					Encryption: Encryption{
						ID:    201,
						Value: hex.EncodeToString([]byte("badKey")),
					},
				},
			},
			wantErr: true,
			wantRepo: CipherKeys{
				Decryption: map[uint8]cipher.AEAD{
					100: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000foo")))

						return key
					}(),
					101: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

						return key
					}(),
				},
				Encryption: Enc{
					ID: 101,
					Value: func() cipher.AEAD {
						key, _ := ParseKey(hex.EncodeToString([]byte("0000000000000bar")))

						return key
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			srv := &Service{
				keyRepo: KeyRepository{
					m:          sync.RWMutex{},
					repository: test.fields.cipherKeys,
				},
			}
			if err := srv.KeysRefresh(test.args.keys); (err != nil) != test.wantErr {
				t.Errorf("Service.KeysRefresh() error = %v, wantErr %v", err, test.wantErr)
			}

			if !reflect.DeepEqual(srv.keyRepo.repository, test.wantRepo) {
				t.Errorf("keys repo not equal, got: %+v\nwant: %+v", srv.keyRepo.repository, test.wantRepo)
			}
		})
	}
}

type MockCipher struct {
	EncryptFn func(cipher.AEAD, []byte) ([]byte, error)
	DecryptFn func(cipher.AEAD, []byte) ([]byte, error)
}

func (c *MockCipher) Encrypt(a cipher.AEAD, b []byte) ([]byte, error) {
	return c.EncryptFn(a, b)
}

func (c *MockCipher) Decrypt(a cipher.AEAD, b []byte) ([]byte, error) {
	return c.DecryptFn(a, b)
}

type MockReader struct {
	ReadFn func(p []byte) (int, error)
}

func (m *MockReader) Read(p []byte) (int, error) {
	return m.ReadFn(p)
}

func TestGCMCipher_Encrypt(t *testing.T) {
	t.Parallel()
	type args struct {
		gcm    cipher.AEAD
		_bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		reader  io.Reader
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				gcm: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
				_bytes: []byte("foo"),
			},
			reader: &MockReader{
				ReadFn: func(p []byte) (int, error) {
					for i := range p {
						p[i] = 0
					}

					return len(p), nil
				},
			},
			want: func() []byte {
				ret, _ := hex.DecodeString(hexString)

				return ret
			}(),
			wantErr: false,
		},
		{
			name: "reader err",
			args: args{
				gcm: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
				_bytes: []byte("foo"),
			},
			reader: &MockReader{
				ReadFn: func(p []byte) (int, error) {
					return 0, ErrRead
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_cipher := &GCMCipher{reader: test.reader}
			got, err := _cipher.Encrypt(test.args.gcm, test.args._bytes)
			if (err != nil) != test.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encrypt() got = %x, want %x", got, test.want)
			}
		})
	}
}

func TestGCMCipher_Decrypt(t *testing.T) {
	t.Parallel()
	type args struct {
		gcm    cipher.AEAD
		_bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				gcm: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
				_bytes: func() []byte {
					ret, _ := hex.DecodeString(hexString)

					return ret
				}(),
			},
			want:    []byte("foo"),
			wantErr: false,
		},
		{
			name: "size err",
			args: args{
				gcm: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
				_bytes: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "gcm err",
			args: args{
				gcm: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
				_bytes: func() []byte {
					ret, _ := hex.DecodeString(hexString)

					return ret[1:]
				}(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_cipher := &GCMCipher{}
			got, err := _cipher.Decrypt(test.args.gcm, test.args._bytes)
			if (err != nil) != test.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decrypt() got = %s, want %s", got, test.want)
			}
		})
	}
}

func TestService_TokenEncryptDecrypt(t *testing.T) {
	t.Parallel()
	service := NewService()
	err := service.KeysRefresh(Keys{
		Decryption: map[uint8]string{
			1: cipherKey,
		},
		Encryption: Encryption{
			ID:    1,
			Value: cipherKey,
		},
	})
	if err != nil {
		t.Errorf("KeysRefresh error: %v", err)
	}

	xidValue := "AAAZU2WSZICQIAZX"

	token, err := service.Encrypt([]byte(xidValue))
	if err != nil {
		t.Errorf("Encrypt error: %v", err)
	}

	decrypted, err := service.Decrypt(token.EncKeyID, token.Value)
	if err != nil {
		t.Errorf("Decrypt error: %v", err)
	}

	if string(decrypted) != xidValue {
		t.Errorf("err")
	}
}
