//go:build linux || freebsd

package daemon

import (
	"testing"

	"github.com/docker/docker/daemon/config"
	"github.com/docker/go-connections/nat"
	containertypes "github.com/moby/moby/api/types/container"
	"gotest.tools/v3/assert"
)

// TestContainerWarningHostAndPublishPorts that a warning is returned when setting network mode to host and specifying published ports.
// This should not be tested on Windows because Windows doesn't support "host" network mode.
func TestContainerWarningHostAndPublishPorts(t *testing.T) {
	testCases := []struct {
		ports    nat.PortMap
		warnings []string
	}{
		{ports: nat.PortMap{}},
		{ports: nat.PortMap{
			"8080": []nat.PortBinding{{HostPort: "8989"}},
		}, warnings: []string{"Published ports are discarded when using host network mode"}},
	}
	muteLogs(t)

	for _, tc := range testCases {
		hostConfig := &containertypes.HostConfig{
			Runtime:      "runc",
			NetworkMode:  "host",
			PortBindings: tc.ports,
		}
		d := &Daemon{}
		cfg, err := config.New()
		assert.NilError(t, err)
		runtimes, err := setupRuntimes(cfg)
		assert.NilError(t, err)
		daemonCfg := &configStore{Config: *cfg, Runtimes: runtimes}
		wrns, err := d.verifyContainerSettings(daemonCfg, hostConfig, &containertypes.Config{}, false)
		assert.NilError(t, err)
		assert.DeepEqual(t, tc.warnings, wrns)
	}
}
