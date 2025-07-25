package daemon

import (
	"testing"

	"github.com/docker/docker/daemon/container"
	containertypes "github.com/moby/moby/api/types/container"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGetInspectData(t *testing.T) {
	c := &container.Container{
		ID:           "inspect-me",
		HostConfig:   &containertypes.HostConfig{},
		State:        container.NewState(),
		ExecCommands: container.NewExecStore(),
	}

	d := &Daemon{
		linkIndex: newLinkIndex(),
	}
	if d.UsesSnapshotter() {
		t.Skip("does not apply to containerd snapshotters, which don't have RWLayer set")
	}
	cfg := &configStore{}
	d.configStore.Store(cfg)

	_, err := d.getInspectData(&cfg.Config, c)
	assert.Check(t, is.ErrorContains(err, "RWLayer of container inspect-me is unexpectedly nil"))

	c.Dead = true
	_, err = d.getInspectData(&cfg.Config, c)
	assert.Check(t, err)
}
