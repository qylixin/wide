package bccsp

import (
	"os"
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
)

var UseGMCrypto bool

var logger = flogging.MustGetLogger("bccsp")

func init() {
	crypt := os.Getenv("BCCSP_CRYPTO_TYPE")
	logger.Debugf("BCCSP_CRYPTO_TYPE: %v", crypt)
	if strings.ToUpper(crypt) == "GM" {
		UseGMCrypto = true
	} else {
		UseGMCrypto = false
	}
	logger.Debugf("UseGMCrypto: %v", UseGMCrypto)
}
