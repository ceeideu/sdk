package crypto

import (
	"crypto/cipher"
	"reflect"
	"sync"
	"testing"
)

const (
	cipherKey = "be2c553d9ea747446fe610c12aab4880027eca7378da8ff9eff3bf1631b3334b"
)

func TestKeyRepository_EncryptionKey(t *testing.T) {
	t.Parallel()
	type fields struct {
		repository CipherKeys
	}
	tests := []struct {
		name    string
		fields  fields
		want    Enc
		wantErr bool
	}{
		{
			name:    "empty",
			fields:  fields{},
			want:    Enc{},
			wantErr: true,
		},
		{
			name: "only decryption in repo",
			fields: fields{
				repository: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						1: func() cipher.AEAD {
							key, _ := ParseKey(cipherKey)

							return key
						}(),
					},
				},
			},
			want:    Enc{},
			wantErr: true,
		},
		{
			name: "both in repo",
			fields: fields{
				repository: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						1: func() cipher.AEAD {
							key, _ := ParseKey(cipherKey)

							return key
						}(),
					},
					Encryption: Enc{
						ID: 2,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(cipherKey)

							return key
						}(),
					},
				},
			},
			want: Enc{
				ID: 2,
				Value: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
			},
		},
		{
			name: "only encryption in repo",
			fields: fields{
				repository: CipherKeys{
					Encryption: Enc{
						ID: 2,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(cipherKey)

							return key
						}(),
					},
				},
			},
			want: Enc{
				ID: 2,
				Value: func() cipher.AEAD {
					key, _ := ParseKey(cipherKey)

					return key
				}(),
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			r := &KeyRepository{
				repository: test.fields.repository,
				m:          sync.RWMutex{},
			}
			got, err := r.EncryptionKey()
			if (err != nil) != test.wantErr {
				t.Errorf("KeyRepository.EncryptionKey() error = %v, wantErr %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("KeyRepository.EncryptionKey() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestKeyRepository_DecryptionKey(t *testing.T) {
	t.Parallel()
	encKey := "be2c553d9ea747446fe610c12aab4880027eca7378da8ff9eff3bf1631b3334b"
	decKey := "da2c553d9ea747446fe610c12aab4880027eca7378da8ff9eff3bf1631b3334b"
	type fields struct {
		repository CipherKeys
	}
	type args struct {
		id uint8
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    cipher.AEAD
		wantErr bool
	}{
		{
			name:    "empty",
			fields:  fields{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "only decryption in repo",
			fields: fields{
				repository: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						1: func() cipher.AEAD {
							key, _ := ParseKey(decKey)

							return key
						}(),
					},
				},
			},
			args: args{id: 1},
			want: func() cipher.AEAD {
				key, _ := ParseKey(decKey)

				return key
			}(),
		},
		{
			name: "both in repo",
			fields: fields{
				repository: CipherKeys{
					Decryption: map[uint8]cipher.AEAD{
						1: func() cipher.AEAD {
							key, _ := ParseKey(decKey)

							return key
						}(),
					},
					Encryption: Enc{
						ID: 2,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(encKey)

							return key
						}(),
					},
				},
			},
			args: args{id: 1},
			want: func() cipher.AEAD {
				key, _ := ParseKey(decKey)

				return key
			}(),
		},
		{
			name: "only encryption in repo",
			fields: fields{
				repository: CipherKeys{
					Encryption: Enc{
						ID: 2,
						Value: func() cipher.AEAD {
							key, _ := ParseKey(encKey)

							return key
						}(),
					},
				},
			},
			args:    args{id: 2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			r := &KeyRepository{
				repository: test.fields.repository,
				m:          sync.RWMutex{},
			}
			got, err := r.DecryptionKey(test.args.id)
			if (err != nil) != test.wantErr {
				t.Errorf("KeyRepository.DecryptionKey() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("KeyRepository.DecryptionKey() = %v, want %v", got, test.want)
			}
		})
	}
}
