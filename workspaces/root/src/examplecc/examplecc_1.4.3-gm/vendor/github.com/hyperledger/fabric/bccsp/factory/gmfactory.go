package factory

import (
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/gm"
	"github.com/pkg/errors"
)

const (
	// GMBasedFactoryName is the name of the factory of the software-based BCCSP implementation
	GMBasedFactoryName = "GM"
)

// GMFactory is the factory of the software-based BCCSP.
type GMFactory struct{}

// Name returns the name of this factory
func (f *GMFactory) Name() string {
	return GMBasedFactoryName
}

// Get returns an instance of BCCSP using Opts.
func (f *GMFactory) Get(config *FactoryOpts) (bccsp.BCCSP, error) {
	// Validate arguments
	if config == nil || config.GmOpts == nil {
		return nil, errors.New("Invalid config. It must not be nil.")
	}

	GmOpts := config.GmOpts

	var ks bccsp.KeyStore
	if GmOpts.Ephemeral == true {
		ks = gm.NewDummyKeyStore()
	} else if GmOpts.FileKeystore != nil {
		fks, err := gm.NewFileBasedKeyStore(nil, GmOpts.FileKeystore.KeyStorePath, false)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize software key store")
		}
		ks = fks
	} else {
		// Default to DummyKeystore
		ks = gm.NewDummyKeyStore()
	}

	return gm.New(ks)
}

// GmOpts contains options for the GMFactory
type GmOpts struct {
	// Keystore Options
	Ephemeral     bool               `mapstructure:"tempkeys,omitempty" json:"tempkeys,omitempty"`
	FileKeystore  *FileKeystoreOpts  `mapstructure:"filekeystore,omitempty" json:"filekeystore,omitempty" yaml:"FileKeyStore"`
	DummyKeystore *DummyKeystoreOpts `mapstructure:"dummykeystore,omitempty" json:"dummykeystore,omitempty"`
}
