package adaptors

import (
	"github.com/davidterranova/envconfig"
)

// EnvconfigAdaptor allows to use envconfig as ConfigReader
type EnvconfigAdaptor struct {
	prefix string
}

// NewEnvConfigAdaptor creates a new adaptor
func NewEnvConfigAdaptor(prefix string) EnvconfigAdaptor {
	return EnvconfigAdaptor{prefix: prefix}
}

// Read satisfies ConfigReader interface
func (a EnvconfigAdaptor) Read(config interface{}) error {
	return envconfig.Process(a.prefix, config)
}
