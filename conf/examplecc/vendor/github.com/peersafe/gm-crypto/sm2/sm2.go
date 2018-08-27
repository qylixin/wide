package sm2

// #cgo LDFLAGS: -L /opt/gopath/src/github.com/hyperledger/fabric/vendor/github.com/peersafe/gm-crypto/usr/lib -lciphersuite_crypto -lciphersuite_smengine -lciphersuite_crypto -lpthread -ldl
// #cgo CFLAGS: -I /opt/gopath/src/github.com/hyperledger/fabric/vendor/github.com/peersafe/gm-crypto/usr/include
// #include <stdio.h>
// #include <string.h>
// #include "openssl/evp.h"
// #include "openssl/bio.h"
// #include "openssl/engine.h"
// #include "openssl/sm2.h"
// #include "openssl/ec.h"
// #include "openssl/bn.h"
// #include "openssl/ossl_typ.h"
// #include "openssl/SMEngine.h"
import "C"
import (
	"crypto"
	"crypto/elliptic"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"unsafe"

	"github.com/peersafe/gm-crypto/engine"
)

var (
	id       = []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
	md       *C.EVP_MD
	curveID  = C.NID_sm2p256v1

        //pubKeyLen  C.int
        //privKeyLen C.int
)

type sm2Signature struct {
	R, S *big.Int
}

// PublicKey define publicKey
type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

// PrivateKey define privateKey
type PrivateKey struct {
	PublicKey
	D *big.Int
}

func GenerateKey() (*PrivateKey, error) {
	pKey := genKey(curveID)
	defer C.EVP_PKEY_free(pKey)

	D := C.EC_KEY_get0_private_key(C.EVP_PKEY_get0_EC_KEY(pKey))
	X := C.BN_new()
	defer C.BN_free(X)
	Y := C.BN_new()
	defer C.BN_free(Y)
	C.EC_POINT_get_affine_coordinates_GFp(
		C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)),
		C.EC_KEY_get0_public_key(C.EVP_PKEY_get0_EC_KEY(pKey)),
		X, Y, nil)

	dBuf := make([]byte, 64)
	xBuf := make([]byte, 64)
	yBuf := make([]byte, 64)

	dLen := C.BN_bn2bin(D, (*C.uchar)(unsafe.Pointer(&dBuf[32])))
	xLen := C.BN_bn2bin(X, (*C.uchar)(unsafe.Pointer(&xBuf[32])))
	yLen := C.BN_bn2bin(Y, (*C.uchar)(unsafe.Pointer(&yBuf[32])))

	Di := big.NewInt(0)
	Di.SetBytes(dBuf[dLen : dLen+32])

	Xi := big.NewInt(0)
	Xi.SetBytes(xBuf[xLen : xLen+32])

	Yi := big.NewInt(0)
	Yi.SetBytes(yBuf[yLen : yLen+32])

	c := P256Sm2()
	pubKey := new(PublicKey)
	pubKey.Curve = c
	pubKey.X = Xi
	pubKey.Y = Yi

	privKey := new(PrivateKey)
	privKey.PublicKey = *pubKey
	privKey.D = Di

	return privKey, nil
}

// Public public key that contains in privateKey
func (p *PrivateKey) Public() crypto.PublicKey {
	return &p.PublicKey
}

// Sign sign the input
func (p *PrivateKey) Sign(rand io.Reader, msg []byte, opts crypto.SignerOpts) ([]byte, error) {
	pKey := genKey(curveID)
	defer C.EVP_PKEY_free(pKey)

	Db := p.D.Bytes()
	D := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Db[0])), C.int(len(Db)), nil)
	defer C.BN_free(D)

	Xb := p.PublicKey.X.Bytes()
	X := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Xb[0])), C.int(len(Xb)), nil)
	defer C.BN_free(X)

	Yb := p.PublicKey.Y.Bytes()
	Y := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Yb[0])), C.int(len(Yb)), nil)
	defer C.BN_free(Y)

	C.EC_KEY_set_private_key(C.EVP_PKEY_get0_EC_KEY(pKey), D)
    point := C.EC_POINT_new(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)))
    C.EC_POINT_set_affine_coordinates_GFp(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)),
                                            point, X, Y, nil)
    C.EC_KEY_set_public_key(C.EVP_PKEY_get0_EC_KEY(pKey), point)
	defer C.EC_POINT_free(point)

	ctx := C.EVP_PKEY_CTX_new(pKey, nil)
	defer C.EVP_PKEY_CTX_free(ctx)

	digest := [64]byte{}
	digestLen := C.size_t(0)

	sig := [128]byte{}
	sigLen := (C.size_t)(C.EVP_PKEY_size(pKey))
    md := C.ENGINE_get_digest((*C.ENGINE)(unsafe.Pointer(engine.Engine)), C.NID_sm3)

	C.SM2_compute_message_digest(md, md, (*C.uchar)(unsafe.Pointer(&msg[0])), C.size_t(len(msg)), (*C.char)(unsafe.Pointer(&id[0])), C.size_t(len(id)), (*C.uchar)(unsafe.Pointer(&digest[0])), &digestLen, C.EVP_PKEY_get0_EC_KEY(C.EVP_PKEY_CTX_get0_pkey(ctx)))
	if C.EVP_PKEY_sign_init(ctx) < 0 {
		log.Println("error: EVP_PKEY_sign_init failed")
		return nil, errors.New("error: EVP_PKEY_sign_init failed")
	}
	//fmt.Printf("digest len = %d, data = %x\n", len(digest), digest)

	if C.EVP_PKEY_sign(ctx, (*C.uchar)(unsafe.Pointer(&sig[0])), &sigLen, (*C.uchar)(unsafe.Pointer(&digest[0])), digestLen) < 0 {
		log.Println("error: EVP_PKEY_sign failed")
		return nil, errors.New("error: EVP_PKEY_sign failed")
	}
    return sig[:sigLen], nil
}

// Verify verify msg
func (p *PublicKey) Verify(msg []byte, sig []byte) bool {
	pKey := genKey(curveID)
	defer C.EVP_PKEY_free(pKey)

	Xb := p.X.Bytes()
	X := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Xb[0])), C.int(len(Xb)), nil)
	defer C.BN_free(X)

	Yb := p.Y.Bytes()
	Y := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Yb[0])), C.int(len(Yb)), nil)
	defer C.BN_free(Y)

	//C.EC_KEY_set_public_key_affine_coordinates(C.EVP_PKEY_get0_EC_KEY(pKey), X, Y)
    point := C.EC_POINT_new(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)))
    C.EC_POINT_set_affine_coordinates_GFp(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)),
                                            point, X, Y, nil)
    C.EC_KEY_set_public_key(C.EVP_PKEY_get0_EC_KEY(pKey), point)
	defer C.EC_POINT_free(point)

	digest := [64]byte{}
	digestLen := C.size_t(0)

	ctx := C.EVP_PKEY_CTX_new(pKey, nil)
	defer C.EVP_PKEY_CTX_free(ctx)

    md := C.ENGINE_get_digest((*C.ENGINE)(unsafe.Pointer(engine.Engine)), C.NID_sm3)
	C.SM2_compute_message_digest(md, md, (*C.uchar)(unsafe.Pointer(&msg[0])), C.size_t(len(msg)), (*C.char)(unsafe.Pointer(&id[0])), C.size_t(len(id)), (*C.uchar)(unsafe.Pointer(&digest[0])), &digestLen, C.EVP_PKEY_get0_EC_KEY(C.EVP_PKEY_CTX_get0_pkey(ctx)))

	if C.EVP_PKEY_verify_init(ctx) < 0 {
		log.Println("error: EVP_PKEY_verify_init failed")
		return false
	}
	//fmt.Printf("digest len = %d, data = %x\n", len(digest), digest)

	if C.EVP_PKEY_verify(ctx, (*C.uchar)(unsafe.Pointer(&sig[0])), (C.size_t)(len(sig)), (*C.uchar)(unsafe.Pointer(&digest[0])), digestLen) != 1 {
		log.Println("error: EVP_PKEY_verify failed")
		//printError()
		return false
	}
	//log.Println("SM2 sig and verify succeed")
	return true
}

//golang stand
func (pub *PublicKey) Encrypt(data []byte) ([]byte, error) {
	return Encrypt(pub, data)
}

func (priv *PrivateKey) Decrypt(data []byte) ([]byte, error) {
	return Decrypt(priv, data)
}

// Encrypt encrypt data
func Encrypt(p *PublicKey, msg []byte) ([]byte, error) {
	pKey := genKey(curveID)
	defer C.EVP_PKEY_free(pKey)

	Xb := p.X.Bytes()
	X := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Xb[0])), C.int(len(Xb)), nil)
	defer C.BN_free(X)

	Yb := p.Y.Bytes()
	Y := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Yb[0])), C.int(len(Yb)), nil)
	defer C.BN_free(Y)

	//C.EC_KEY_set_public_key_affine_coordinates(C.EVP_PKEY_get0_EC_KEY(pKey), X, Y)
    point := C.EC_POINT_new(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)))
    C.EC_POINT_set_affine_coordinates_GFp(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)),
                                            point, X, Y, nil)
    C.EC_KEY_set_public_key(C.EVP_PKEY_get0_EC_KEY(pKey), point)
	defer C.EC_POINT_free(point)

	ctx := C.EVP_PKEY_CTX_new(pKey, nil)
	defer C.EVP_PKEY_CTX_free(ctx)

	ciphertext := [512]byte{}
	mlen := (C.size_t)(len(msg))
	clen := (C.size_t)(len(ciphertext))

	if C.EVP_PKEY_encrypt_init(ctx) != 1 {
		log.Println("error: EVP_PKEY_encrypt_init")
		return nil, errors.New("error: EVP_PKEY_encrypt_init failed")
	}

	if C.EVP_PKEY_encrypt(ctx, (*C.uchar)(unsafe.Pointer(&ciphertext[0])), &clen, (*C.uchar)(unsafe.Pointer(&msg[0])), mlen) <= 0 {
		log.Println("error: EVP_PKEY_encrypt")
		return nil, errors.New("error: EVP_PKEY_encrypt failed")
	}

	out := (ciphertext)
	return out[:clen], nil
}

// Decrypt decrypt data
func Decrypt(p *PrivateKey, msg []byte) ([]byte, error) {
	pKey := genKey(curveID)
	defer C.EVP_PKEY_free(pKey)

	Db := p.D.Bytes()
	D := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Db[0])), C.int(len(Db)), nil)
	defer C.BN_free(D)

	Xb := p.PublicKey.X.Bytes()
	X := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Xb[0])), C.int(len(Xb)), nil)
	defer C.BN_free(X)

	Yb := p.PublicKey.Y.Bytes()
	Y := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Yb[0])), C.int(len(Yb)), nil)
	defer C.BN_free(Y)

	C.EC_KEY_set_private_key(C.EVP_PKEY_get0_EC_KEY(pKey), D)
	//C.EC_KEY_set_public_key_affine_coordinates(C.EVP_PKEY_get0_EC_KEY(pKey), X, Y)
    point := C.EC_POINT_new(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)))
    C.EC_POINT_set_affine_coordinates_GFp(C.EC_KEY_get0_group(C.EVP_PKEY_get0_EC_KEY(pKey)),
                                            point, X, Y, nil)
    C.EC_KEY_set_public_key(C.EVP_PKEY_get0_EC_KEY(pKey), point)
	defer C.EC_POINT_free(point)

	ctx := C.EVP_PKEY_CTX_new(pKey, nil)
	defer C.EVP_PKEY_CTX_free(ctx)

	plain := [512]byte{}
	plen := (C.size_t)(len(plain))
	clen := (C.size_t)(len(msg))

	if C.EVP_PKEY_decrypt_init(ctx) != 1 {
		log.Println("error: EVP_PKEY_decrypt_init")
		return nil, errors.New("error: EVP_PKEY_decrypt_init failed")
	}

	if C.EVP_PKEY_decrypt(ctx, (*C.uchar)(unsafe.Pointer(&plain[0])), &plen, (*C.uchar)(unsafe.Pointer(&msg[0])), clen) <= 0 {
		log.Println("error: EVP_PKEY_decrypt")
		return nil, errors.New("error: EVP_PKEY_decrypt failed")
	}

	out := (plain)
	return out[:plen], nil
}

func Sign(p *PrivateKey, msg []byte) (r, s *big.Int, err error) {
    sig, err := p.Sign(nil, msg, nil)
    if err != nil {
        return nil, nil, err
    }

    sdata := C.ECDSA_SIG_new()
    defer C.ECDSA_SIG_free(sdata)

    sigper := uintptr(unsafe.Pointer(&sig[0]))
    if C.d2i_ECDSA_SIG(&sdata, (**C.uchar)(unsafe.Pointer(&sigper)), (C.long)(len(sig))) == nil {
		log.Println("error: d2i_ECDSA_SIG failed")
		return nil, nil, errors.New("error: d2i_ECDSA_SIG failed")
	}
	var pr *C.BIGNUM
	var ps *C.BIGNUM
	C.ECDSA_SIG_get0(sdata, &pr, &ps)

	rBuf := make([]byte, 64)
	sBuf := make([]byte, 64)

	rLen := C.BN_bn2bin(pr, (*C.uchar)(unsafe.Pointer(&rBuf[32])))
	sLen := C.BN_bn2bin(ps, (*C.uchar)(unsafe.Pointer(&sBuf[32])))
	//fmt.Printf("rlen = %d, data = %x\n", rLen, rBuf)
	//fmt.Printf("slen = %d, data = %x\n", sLen, sBuf)

	r = big.NewInt(0)
	r.SetBytes(rBuf[rLen : rLen+32])

	s = big.NewInt(0)
	s.SetBytes(sBuf[sLen : sLen+32])

	return r, s, nil
}

func Verify(p *PublicKey, msg []byte, r, s *big.Int) bool {
    sdata := C.ECDSA_SIG_new()
    defer C.ECDSA_SIG_free(sdata)

	Rb := r.Bytes()
	R := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Rb[0])), C.int(len(Rb)), nil)
	//defer C.BN_free(R)

	Sb := s.Bytes()
	S := C.BN_bin2bn((*C.uchar)(unsafe.Pointer(&Sb[0])), C.int(len(Sb)), nil)
	//defer C.BN_free(S)

	C.ECDSA_SIG_set0(sdata, R, S)

	sig := [256]byte{}
	sigper := uintptr(unsafe.Pointer(&sig[0]))
	sigLen := C.i2d_ECDSA_SIG(sdata, (**C.uchar)(unsafe.Pointer(&sigper)))

    return p.Verify(msg, sig[:sigLen])
}

func genKey(curveID int) *C.EVP_PKEY {
	var ret *C.EVP_PKEY
	ctx := C.EVP_PKEY_CTX_new_id(C.EVP_PKEY_EC, nil)
	defer C.EVP_PKEY_CTX_free(ctx)

	if ctx == nil {
		log.Println("error: EVP_PKEY_CTX_new_id failed")
		return nil
	}
	if C.EVP_PKEY_keygen_init(ctx) < 0 {
		log.Println("error: EVP_PKEY_keygen_init failed")
		return nil
	}
	if C.EVP_PKEY_CTX_ctrl(ctx, C.EVP_PKEY_EC, C.EVP_PKEY_OP_PARAMGEN|C.EVP_PKEY_OP_KEYGEN, C.EVP_PKEY_CTRL_EC_PARAMGEN_CURVE_NID, C.int(curveID), nil) < 0 {
		log.Println("error: EVP_PKEY_CTX_set_ec_paramgen_curve_nid failed")
		return nil
	}
	if C.EVP_PKEY_keygen(ctx, &ret) < 0 {
		log.Println("error: EVP_PKEY_keygen failed")
		freeEvpKey(ret)
		return nil
	}
	return ret
}

// Sign signs a hash (which should be the result of hashing a larger message)
// using the private key, priv. If the hash is longer than the bit-length of the
// private key's curve order, the hash will be truncated to that length.  It
// returns the signature as a pair of integers. The security of the private key
// depends on the entropy of rand.
// func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
// }

func freeEvpKey(ret *C.EVP_PKEY) {
	if ret != nil {
		C.EVP_PKEY_free(ret)
		ret = nil
	}
}

func printError() {
	fuc := make([]byte, 100)
	line := 0
	for {
		num := C.ERR_get_error_line((**C.char)(unsafe.Pointer(&fuc)), (*C.int)(unsafe.Pointer(&line)))
		fmt.Printf("ERR_get_error_line func: %s, line = %d.\n", C.GoString((*C.char)(unsafe.Pointer(&fuc))), line)
		fmt.Printf("ERR_get_error_line num = %d.\n", num)
		if num == 0 {
			break
		}
	}
}
