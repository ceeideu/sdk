package crypto

import (
	"crypto/cipher"
	"errors"
	"sync"
)

var (
	ErrEncrypt       = errors.New("encrypt error")
	ErrNilEncKey     = errors.New("nil encrypt key error")
	ErrKeyNotPresent = errors.New("key not present")
)

type CipherKeys struct {
	Decryption map[uint8]cipher.AEAD
	Encryption Enc
}

type Enc struct {
	ID    uint8
	Value cipher.AEAD
}

type KeyRepository struct {
	repository CipherKeys

	m sync.RWMutex
}

func (r *KeyRepository) EncryptionKey() (Enc, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.repository.Encryption.Value == nil {
		return Enc{}, ErrNilEncKey
	}

	return r.repository.Encryption, nil
}

func (r *KeyRepository) DecryptionKey(id uint8) (cipher.AEAD, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if k, ok := r.repository.Decryption[id]; ok {
		return k, nil
	}

	return nil, ErrKeyNotPresent
}

func (r *KeyRepository) Set(ck CipherKeys) {
	r.m.Lock()
	defer r.m.Unlock()
	r.repository = ck
}
