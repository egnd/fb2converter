// Package state defines stared program state.
package state

import (
	"go.uber.org/zap"

	"github.com/rupor-github/fb2converter/config"
)

// LocalEnv keeps everything program needs in a single place.
type LocalEnv struct {
	Debug string
	Mhl   int

	Cfg *config.Config
	Log *zap.Logger
}

// NewLocalEnv creates LocalEnv and initializes it.
func NewLocalEnv() *LocalEnv {
	return &LocalEnv{}
}

// In "github.com/urfave/cli" the only way I found to share state between "app" and "command" without global variables
// is to use hidden GenericFlag. To implement the mechanics we need following code...
const (
	FlagName = "$-localenv-$"
)

// Set implements cli's flag interface
func (e *LocalEnv) Set(value string) error {
	panic("localenv value should never be set directly")
}

// String implements cli's flag interface
func (e *LocalEnv) String() string {
	return "local-env"
}
