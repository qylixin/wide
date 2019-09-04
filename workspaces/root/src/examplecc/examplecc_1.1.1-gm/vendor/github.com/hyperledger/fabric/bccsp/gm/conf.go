package gm

import (
	"crypto/elliptic"
	"crypto/sm3"
	"hash"
)

type config struct {
	ellipticCurve elliptic.Curve
	hashFunction  func() hash.Hash
}

func (conf *config) setSecurityLevel() (err error) {
	conf.ellipticCurve = elliptic.P256SM2()
	conf.hashFunction = sm3.New
	return nil
}
