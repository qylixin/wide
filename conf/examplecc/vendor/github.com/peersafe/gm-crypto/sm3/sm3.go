package sm3

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
// const int x = 42;
import "C"

import (
	"hash"
	"unsafe"

	"github.com/peersafe/gm-crypto/engine"
)

// Sm3Size the size of sm3 digest
const Sm3Size = 32

// Sm3BlockSize the size of sm3 block
const Sm3BlockSize = 64

type digest struct {
	engine    *C.ENGINE
	msgDigest *C.EVP_MD_CTX
}

type SM3 struct {
	digest      [8]uint32     // digest represents the partial evaluation of V
	length      uint64        // length of the message
	unhandleMsg []byte        // uint8  //
	mdctx       *C.EVP_MD_CTX // *EVP_MD_CTX //
}

// New returns a new hash.Hash computing the SM3 checksum
func New() hash.Hash {
	var sm3 SM3

	sm3.Reset()
	return &sm3
}

func (sm3 *SM3) Reset() {
	// Reset digest
	sm3.digest[0] = 0x7380166f
	sm3.digest[1] = 0x4914b2b9
	sm3.digest[2] = 0x172442d7
	sm3.digest[3] = 0xda8a0600
	sm3.digest[4] = 0xa96f30bc
	sm3.digest[5] = 0x163138aa
	sm3.digest[6] = 0xe38dee4d
	sm3.digest[7] = 0xb0fb0e4e

	sm3.length = 0 // Reset numberic states
	sm3.unhandleMsg = nil
}

func (sm3 *SM3) BlockSize() int { return Sm3BlockSize }

func (sm3 *SM3) Size() int { return Sm3Size }

func (sm3 *SM3) Write(msg []byte) (nn int, err error) {
	if msg == nil {
		return 0, nil
	}

	sm3.unhandleMsg = append(sm3.unhandleMsg, msg...)
	return len(msg), nil
}

func (sm3 *SM3) checkSum() [Sm3Size]byte {
	var digest [Sm3Size]byte
	return digest
}

func (sm3 *SM3) Sum(msg []byte) []byte {
	if msg == nil && sm3.unhandleMsg == nil {
		return nil
	}

	msg = append(sm3.unhandleMsg, msg...)

	out := [Sm3Size]byte{}
	size := C.uint(Sm3Size)

	evpMd := C.ENGINE_get_digest((*C.ENGINE)(unsafe.Pointer(engine.Engine)), C.NID_sm3)
	sm3.mdctx = C.EVP_MD_CTX_new()
	C.EVP_DigestInit_ex(sm3.mdctx, evpMd, (*C.ENGINE)(unsafe.Pointer(engine.Engine)))

	C.EVP_DigestUpdate(sm3.mdctx, unsafe.Pointer(&msg[0]), C.size_t(len(msg)))

	C.EVP_DigestFinal_ex(sm3.mdctx, (*C.uchar)(unsafe.Pointer(&out[0])), &size)

	C.EVP_MD_CTX_free(sm3.mdctx)

	sm3.unhandleMsg = nil

	return out[:size]
}

func Sm3Sum(data []byte) []byte {
	var sm3 SM3

	sm3.Reset()
	sm3.Write(data)
	return sm3.Sum(nil)
}
