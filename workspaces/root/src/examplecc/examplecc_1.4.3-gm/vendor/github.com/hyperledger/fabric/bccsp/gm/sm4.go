package gm

import (
	"crypto/rand"
	"errors"
	"fmt"

	"crypto/sm4"

	"github.com/hyperledger/fabric/bccsp"
)

// GetRandomBytes returns len random looking bytes
func GetRandomBytes(len int) ([]byte, error) {
	if len < 0 {
		return nil, errors.New("Len must be larger than 0")
	}

	buffer := make([]byte, len)

	n, err := rand.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n != len {
		return nil, fmt.Errorf("Buffer not filled. Requested [%d], got [%d]", len, n)
	}

	return buffer, nil
}

type sm4Encryptor struct{}

func (e *sm4Encryptor) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) ([]byte, error) {
	ciphertext := make([]byte, sm4.GetChiperPadLength(len(plaintext)))
	err := sm4.EncryptWithPad(k.(*sm4PrivateKey).privKey, ciphertext, plaintext)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

type sm4Decryptor struct{}

func (*sm4Decryptor) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) ([]byte, error) {
	plaintext := make([]byte, len(ciphertext))
	finalLength, err := sm4.DecryptWithPad(k.(*sm4PrivateKey).privKey, plaintext, ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext[:finalLength], nil
}
