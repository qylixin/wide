package gm

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/utils"
)

func signSM2(k *ecdsa.PrivateKey, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, k, digest)
	if err != nil {
		return nil, err
	}

	//s, _, err = utils.ToLowS(&k.PublicKey, s)
	//if err != nil {
	//	return nil, err
	//}

	return utils.MarshalECDSASignature(r, s)
}

func verifySM2(k *ecdsa.PublicKey, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	r, s, err := utils.UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}

	//lowS, err := utils.IsLowS(k, s)
	//if err != nil {
	//	return false, err
	//}
	//
	//if !lowS {
	//	return false, fmt.Errorf("Invalid S. Must be smaller than half the order [%s][%s].", s, utils.GetCurveHalfOrdersAt(k.Curve))
	//}

	return ecdsa.Verify(k, digest, r, s), nil
}

type sm2Signer struct{}

func (s *sm2Signer) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	return signSM2(k.(*sm2PrivateKey).privKey, digest, opts)
}

type sm2PrivateKeyVerifier struct{}

func (v *sm2PrivateKeyVerifier) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	return verifySM2(&(k.(*sm2PrivateKey).privKey.PublicKey), signature, digest, opts)
}

type sm2PublicKeyVerifier struct{}

func (v *sm2PublicKeyVerifier) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	return verifySM2(k.(*sm2PublicKey).pubKey, signature, digest, opts)
}
