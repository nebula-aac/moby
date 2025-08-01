// Code generated by go-swagger; DO NOT EDIT.

package types

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/moby/moby/api/types/plugin"
)

// Plugin A plugin for the Engine API
//
// swagger:model Plugin
type Plugin struct {

	// config
	// Required: true
	Config PluginConfig `json:"Config"`

	// True if the plugin is running. False if the plugin is not running, only installed.
	// Example: true
	// Required: true
	Enabled bool `json:"Enabled"`

	// Id
	// Example: 5724e2c8652da337ab2eedd19fc6fc0ec908e4bd907c7421bf6a8dfc70c4c078
	ID string `json:"Id,omitempty"`

	// name
	// Example: tiborvass/sample-volume-plugin
	// Required: true
	Name string `json:"Name"`

	// plugin remote reference used to push/pull the plugin
	// Example: localhost:5000/tiborvass/sample-volume-plugin:latest
	PluginReference string `json:"PluginReference,omitempty"`

	// settings
	// Required: true
	Settings PluginSettings `json:"Settings"`
}

// PluginConfig The config of a plugin.
//
// swagger:model PluginConfig
type PluginConfig struct {

	// args
	// Required: true
	Args PluginConfigArgs `json:"Args"`

	// description
	// Example: A sample volume plugin for Docker
	// Required: true
	Description string `json:"Description"`

	// Docker Version used to create the plugin
	// Example: 17.06.0-ce
	DockerVersion string `json:"DockerVersion,omitempty"`

	// documentation
	// Example: https://docs.docker.com/engine/extend/plugins/
	// Required: true
	Documentation string `json:"Documentation"`

	// entrypoint
	// Example: ["/usr/bin/sample-volume-plugin","/data"]
	// Required: true
	Entrypoint []string `json:"Entrypoint"`

	// env
	// Example: [{"Description":"If set, prints debug messages","Name":"DEBUG","Settable":null,"Value":"0"}]
	// Required: true
	Env []PluginEnv `json:"Env"`

	// interface
	// Required: true
	Interface PluginConfigInterface `json:"Interface"`

	// ipc host
	// Example: false
	// Required: true
	IpcHost bool `json:"IpcHost"`

	// linux
	// Required: true
	Linux PluginConfigLinux `json:"Linux"`

	// mounts
	// Required: true
	Mounts []PluginMount `json:"Mounts"`

	// network
	// Required: true
	Network PluginConfigNetwork `json:"Network"`

	// pid host
	// Example: false
	// Required: true
	PidHost bool `json:"PidHost"`

	// propagated mount
	// Example: /mnt/volumes
	// Required: true
	PropagatedMount string `json:"PropagatedMount"`

	// user
	User PluginConfigUser `json:"User,omitempty"`

	// work dir
	// Example: /bin/
	// Required: true
	WorkDir string `json:"WorkDir"`

	// rootfs
	Rootfs *PluginConfigRootfs `json:"rootfs,omitempty"`
}

// PluginConfigArgs plugin config args
//
// swagger:model PluginConfigArgs
type PluginConfigArgs struct {

	// description
	// Example: command line arguments
	// Required: true
	Description string `json:"Description"`

	// name
	// Example: args
	// Required: true
	Name string `json:"Name"`

	// settable
	// Required: true
	Settable []string `json:"Settable"`

	// value
	// Required: true
	Value []string `json:"Value"`
}

// PluginConfigInterface The interface between Docker and the plugin
//
// swagger:model PluginConfigInterface
type PluginConfigInterface struct {

	// Protocol to use for clients connecting to the plugin.
	// Example: some.protocol/v1.0
	// Enum: ["","moby.plugins.http/v1"]
	ProtocolScheme string `json:"ProtocolScheme,omitempty"`

	// socket
	// Example: plugins.sock
	// Required: true
	Socket string `json:"Socket"`

	// types
	// Example: ["docker.volumedriver/1.0"]
	// Required: true
	Types []plugin.CapabilityID `json:"Types"`
}

// PluginConfigLinux plugin config linux
//
// swagger:model PluginConfigLinux
type PluginConfigLinux struct {

	// allow all devices
	// Example: false
	// Required: true
	AllowAllDevices bool `json:"AllowAllDevices"`

	// capabilities
	// Example: ["CAP_SYS_ADMIN","CAP_SYSLOG"]
	// Required: true
	Capabilities []string `json:"Capabilities"`

	// devices
	// Required: true
	Devices []PluginDevice `json:"Devices"`
}

// PluginConfigNetwork plugin config network
//
// swagger:model PluginConfigNetwork
type PluginConfigNetwork struct {

	// type
	// Example: host
	// Required: true
	Type string `json:"Type"`
}

// PluginConfigRootfs plugin config rootfs
//
// swagger:model PluginConfigRootfs
type PluginConfigRootfs struct {

	// diff ids
	// Example: ["sha256:675532206fbf3030b8458f88d6e26d4eb1577688a25efec97154c94e8b6b4887","sha256:e216a057b1cb1efc11f8a268f37ef62083e70b1b38323ba252e25ac88904a7e8"]
	DiffIds []string `json:"diff_ids"`

	// type
	// Example: layers
	Type string `json:"type,omitempty"`
}

// PluginConfigUser plugin config user
//
// swagger:model PluginConfigUser
type PluginConfigUser struct {

	// g ID
	// Example: 1000
	GID uint32 `json:"GID,omitempty"`

	// UID
	// Example: 1000
	UID uint32 `json:"UID,omitempty"`
}

// PluginSettings Settings that can be modified by users.
//
// swagger:model PluginSettings
type PluginSettings struct {

	// args
	// Required: true
	Args []string `json:"Args"`

	// devices
	// Required: true
	Devices []PluginDevice `json:"Devices"`

	// env
	// Example: ["DEBUG=0"]
	// Required: true
	Env []string `json:"Env"`

	// mounts
	// Required: true
	Mounts []PluginMount `json:"Mounts"`
}
