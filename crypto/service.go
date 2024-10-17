package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrRead      = errors.New("read error")
	ErrNonceSize = errors.New("nonce size error")
	ErrOpen      = errors.New("open error")
	ErrCipher    = errors.New("cipher error")
	ErrDecode    = errors.New("client decode error")
	ErrBlock     = errors.New("block error")
)

type Service struct {
	cipher  Cipher
	keyRepo KeyRepository
}

type Cipher interface {
	Encrypt(aead cipher.AEAD, _bytes []byte) ([]byte, error)
	Decrypt(aead cipher.AEAD, _bytes []byte) ([]byte, error)
}

func NewService() *Service {
	return &Service{
		cipher: &GCMCipher{reader: rand.Reader},
	}
}

type EncResp struct {
	Value    []byte
	EncKeyID uint8
}

func (s *Service) Encrypt(data []byte) (EncResp, error) {
	encryptionKey, err := s.keyRepo.EncryptionKey()
	if err != nil {
		return EncResp{}, err
	}
	encrypted, err := s.cipher.Encrypt(encryptionKey.Value, data)
	if err != nil {
		return EncResp{}, fmt.Errorf("%w: %w", ErrEncrypt, err)
	}

	return EncResp{
		Value:    encrypted,
		EncKeyID: encryptionKey.ID,
	}, nil
}

func (s *Service) Decrypt(keyID uint8, data []byte) ([]byte, error) {
	key, err := s.keyRepo.DecryptionKey(keyID)
	if err != nil {
		return nil, err
	}

	decrypted, err := s.cipher.Decrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "decrypt error", err)
	}

	return decrypted, nil
}

type Keys struct {
	Decryption map[uint8]string
	Encryption Encryption
}

type Encryption struct {
	ID    uint8
	Value string
}

func (s *Service) KeysRefresh(keys Keys) error {
	repo := CipherKeys{
		Decryption: map[uint8]cipher.AEAD{},
		Encryption: Enc{},
	}
	for _id, dkey := range keys.Decryption {
		key, err := ParseKey(dkey)
		if err != nil {
			return err
		}

		repo.Decryption[_id] = key
	}

	key, err := ParseKey(keys.Encryption.Value)
	if err != nil {
		return err
	}

	repo.Encryption = Enc{
		ID:    keys.Encryption.ID,
		Value: key,
	}
	s.keyRepo.Set(repo)

	return nil
}

func ParseKey(value string) (cipher.AEAD, error) {
	_bytes, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	block, err := aes.NewCipher(_bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCipher, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBlock, err)
	}

	return gcm, nil
}

type GCMCipher struct {
	reader io.Reader
}

func NewGCMCipher() *GCMCipher {
	return &GCMCipher{reader: rand.Reader}
}

func (c *GCMCipher) Encrypt(gcm cipher.AEAD, _bytes []byte) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(c.reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRead, err)
	}

	return gcm.Seal(nonce, nonce, _bytes, nil), nil
}

func (c *GCMCipher) Decrypt(gcm cipher.AEAD, _bytes []byte) ([]byte, error) {
	nonceSize := gcm.NonceSize()

	if len(_bytes) < nonceSize {
		return nil, ErrNonceSize
	}

	nonce, ciphertext := _bytes[:nonceSize], _bytes[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpen, err)
	}

	return plaintext, nil
}
