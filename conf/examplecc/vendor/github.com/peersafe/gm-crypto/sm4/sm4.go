package sm4

// #cgo LDFLAGS: -L /opt/gopath/src/github.com/hyperledger/fabric/vendor/github.com/peersafe/gm-crypto/usr/lib -lciphersuite_crypto -lciphersuite_smengine -lciphersuite_crypto -lpthread -ldl
// #cgo CFLAGS: -I /opt/gopath/src/github.com/hyperledger/fabric/vendor/github.com/peersafe/gm-crypto/usr/include
// #include <stdio.h>
// #include <string.h>
// #include "openssl/engine.h"
// #include "openssl/evp.h"
// #include "openssl/hmac.h"
// #include "openssl/rand.h"
// #include "openssl/ossl_typ.h"
// #include "openssl/obj_mac.h"
import "C"

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/peersafe/gm-crypto/engine"
)

const BlockSize = 16
const default_iv = "1234567890123456"

type SM4Key []byte

type Sm4Cipher struct {
	subkeys []uint32
	block1  []uint32
	block2  []byte
	key     []byte
	iv      []byte
}

func ReadKeyFromMem(data []byte, pwd []byte) (SM4Key, error) {
	block, _ := pem.Decode(data)
	if x509.IsEncryptedPEMBlock(block) {
		if block.Type != "SM4 ENCRYPTED KEY" {
			return nil, errors.New("SM4: unknown type")
		}
		if pwd == nil {
			return nil, errors.New("SM4: need passwd")
		}
		data, err := x509.DecryptPEMBlock(block, pwd)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	if block.Type != "SM4 KEY" {
		return nil, errors.New("SM4: unknown type")
	}
	return block.Bytes, nil
}

func ReadKeyFromPem(FileName string, pwd []byte) (SM4Key, error) {
	data, err := ioutil.ReadFile(FileName)
	if err != nil {
		return nil, err
	}
	return ReadKeyFromMem(data, pwd)
}

func WriteKeytoMem(key SM4Key, pwd []byte) ([]byte, error) {
	if pwd != nil {
		block, err := x509.EncryptPEMBlock(rand.Reader,
			"SM4 ENCRYPTED KEY", key, pwd, x509.PEMCipherAES256)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(block), nil
	} else {
		block := &pem.Block{
			Type:  "SM4 KEY",
			Bytes: key,
		}
		return pem.EncodeToMemory(block), nil
	}
}

func WriteKeyToPem(FileName string, key SM4Key, pwd []byte) (bool, error) {
	var block *pem.Block

	if pwd != nil {
		var err error
		block, err = x509.EncryptPEMBlock(rand.Reader,
			"SM4 ENCRYPTED KEY", key, pwd, x509.PEMCipherAES256)
		if err != nil {
			return false, err
		}
	} else {
		block = &pem.Block{
			Type:  "SM4 KEY",
			Bytes: key,
		}
	}
	file, err := os.Create(FileName)
	if err != nil {
		return false, err
	}
	defer file.Close()
	err = pem.Encode(file, block)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func NewDefaultCipher(key []byte) (*Sm4Cipher, error) {
	c := new(Sm4Cipher)
	c.key = key
	c.iv = []byte(default_iv)
	return c, nil
}

// NewCipher creates and returns a new cipher.Block.
func NewCipher(key, iv []byte) (*Sm4Cipher, error) {
	c := new(Sm4Cipher)
	c.key = key
	c.iv = iv
	return c, nil
}

func (c *Sm4Cipher) BlockSize() int {
	return BlockSize
}

func Encrypt(key, dst, src []byte) int32 {
	sm4Cipher, _ := NewDefaultCipher(key)
	return sm4Cipher.Encrypt(dst, src)
}

//修改后的加密核心函数
func (c *Sm4Cipher) Encrypt(dst, src []byte) int32 {
	tmplen := (C.int)(0)
	clen := (C.int)(0)

	cipher := C.ENGINE_get_cipher((*C.ENGINE)(unsafe.Pointer(engine.Engine)), C.NID_sm4_ctr)
	ctx := C.EVP_CIPHER_CTX_new()
	defer C.EVP_CIPHER_CTX_free(ctx)

	C.EVP_EncryptInit_ex(ctx, cipher, (*C.ENGINE)(unsafe.Pointer(engine.Engine)), (*C.uchar)(unsafe.Pointer(&c.key[0])), (*C.uchar)(unsafe.Pointer(&c.iv[0])))
	C.EVP_EncryptUpdate(ctx, (*C.uchar)(unsafe.Pointer(&dst[0])), &clen, (*C.uchar)(unsafe.Pointer(&src[0])), (C.int)(len(src)))
	//fmt.Printf("s0 = %x\n", s0)
	//fmt.Printf("d0-%d = %x\n", i/16, d0[:16])
	//fmt.Printf("clen = %d\n", clen)
	C.EVP_EncryptFinal_ex(ctx, (*C.uchar)(unsafe.Pointer(&dst[clen-1])), &tmplen)
	//fmt.Printf("tmplen = %d\n", tmplen)
	return (int32)(clen + tmplen)
}

func Decrypt(key, dst, src []byte) int32 {
	sm4Cipher, _ := NewDefaultCipher(key)
	return sm4Cipher.Decrypt(dst, src)
}

//修改后的解密核心函数
func (c *Sm4Cipher) Decrypt(dst, src []byte) int32 {
	tmplen := (C.int)(0)
	plen := (C.int)(0)

	cipher := C.ENGINE_get_cipher((*C.ENGINE)(unsafe.Pointer(engine.Engine)), C.NID_sm4_ctr)
	ctx := C.EVP_CIPHER_CTX_new()
	defer C.EVP_CIPHER_CTX_free(ctx)

	C.EVP_DecryptInit_ex(ctx, cipher, (*C.ENGINE)(unsafe.Pointer(engine.Engine)), (*C.uchar)(unsafe.Pointer(&c.key[0])), (*C.uchar)(unsafe.Pointer(&c.iv[0])))
	C.EVP_DecryptUpdate(ctx, (*C.uchar)(unsafe.Pointer(&dst[0])), &plen, (*C.uchar)(unsafe.Pointer(&src[0])), (C.int)(len(src)))
	//fmt.Printf("plen = %d\n", plen)
	//fmt.Printf("d1-%d = %x\n", i/16, d1)
	C.EVP_DecryptFinal_ex(ctx, (*C.uchar)(unsafe.Pointer(&dst[plen-1])), &tmplen)
	//fmt.Printf("tmplen = %d\n", tmplen)
	//fmt.Printf("d1-%d = %x\n", i/16, dst[i:])
	return (int32)(plen + tmplen)
}
