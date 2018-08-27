package engine

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
// #include "openssl/SMEngine.h"
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

const (
	ENGINEID                     = "CipherSuite_SM"
	OPENSSL_INIT_ADD_ALL_CIPHERS = C.OPENSSL_INIT_ADD_ALL_CIPHERS
	OPENSSL_INIT_ADD_ALL_DIGESTS = C.OPENSSL_INIT_ADD_ALL_DIGESTS
	OPENSSL_INIT_LOAD_CONFIG     = C.OPENSSL_INIT_LOAD_CONFIG
	OPENSSL_INIT_ENGINE_DYNAMIC  = C.OPENSSL_INIT_ENGINE_DYNAMIC
)

var Engine *C.ENGINE

func ENGINE_init(engine *C.ENGINE) int {
	return int(C.ENGINE_init(engine))
}

// Open engine for sm2 sm3 sm4
func init() {
	var err error

	C.OPENSSL_init_crypto(OPENSSL_INIT_ADD_ALL_CIPHERS|OPENSSL_INIT_ADD_ALL_DIGESTS, nil)
	C.ENGINE_load_CipherSuite()
	eid := []byte(ENGINEID)
	Engine = C.ENGINE_by_id((*C.char)(unsafe.Pointer(&eid[0])))
	if Engine == nil {
		log.Println(fmt.Sprintf("SM Engine is NULL."))
		panic(err)
	}
	ENGINE_init(Engine)
	C.ENGINE_register_complete(Engine)
}
