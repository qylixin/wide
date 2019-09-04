package bccsp

const (
	SM2 = "SM2"
	SM3 = "SM3"
	SM4 = "SM4"
)

// ECDSAP256KeyGenOpts contains options for ECDSA key generation with curve P-256.
type SM2KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *SM2KeyGenOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM2KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM4KeyGenOpts contains options for SM4 key generation at 128 security level
type SM4KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *SM4KeyGenOpts) Algorithm() string {
	return SM4
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM4KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM4ImportKeyOpts contains options for importing AES 256 keys.
type SM4ImportKeyOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM4ImportKeyOpts) Algorithm() string {
	return SM4
}

// Ephemeral returns true if the key generated has to be ephemeral,
// false otherwise.
func (opts *SM4ImportKeyOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM2PKIXPublicKeyImportOpts contains options for SM2 public key importation in PKIX format
type SM2PKIXPublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2PKIXPublicKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM2PKIXPublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM2PrivateKeyImportOpts contains options for ECDSA secret key importation in DER format
// or PKCS#8 format.
type SM2PrivateKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2PrivateKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM2PrivateKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM2GoPublicKeyImportOpts contains options for ECDSA key importation from ecdsa.PublicKey
type SM2GoPublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2GoPublicKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM2GoPublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}
