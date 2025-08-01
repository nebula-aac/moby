//go:build unix

package command

import (
	"testing"

	"github.com/moby/moby/v2/daemon/config"
	"github.com/spf13/pflag"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestDaemonParseShmSize(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	conf, err := config.New()
	assert.NilError(t, err)
	installConfigFlags(conf, flags)
	// By default `--default-shm-size=64M`
	assert.Check(t, is.Equal(int64(64*1024*1024), conf.ShmSize.Value()))
	assert.Check(t, flags.Set("default-shm-size", "128M"))
	assert.Check(t, is.Equal(int64(128*1024*1024), conf.ShmSize.Value()))
}
