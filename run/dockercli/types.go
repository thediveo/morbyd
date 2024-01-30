// Source: https://github.com/docker/cli/blob/v25.0.1/cli/compose/types/types.go
//
// Apache License 2.0, https://github.com/docker/cli/blob/master/LICENSE
//
// Taken only type ServiceVolumeConfig definition, with direct dependency types.

package dockercli

// ServiceVolumeConfig are references to a volume used by a service
type ServiceVolumeConfig struct {
	Type        string                `yaml:",omitempty" json:"type,omitempty"`
	Source      string                `yaml:",omitempty" json:"source,omitempty"`
	Target      string                `yaml:",omitempty" json:"target,omitempty"`
	ReadOnly    bool                  `mapstructure:"read_only" yaml:"read_only,omitempty" json:"read_only,omitempty"`
	Consistency string                `yaml:",omitempty" json:"consistency,omitempty"`
	Bind        *ServiceVolumeBind    `yaml:",omitempty" json:"bind,omitempty"`
	Volume      *ServiceVolumeVolume  `yaml:",omitempty" json:"volume,omitempty"`
	Tmpfs       *ServiceVolumeTmpfs   `yaml:",omitempty" json:"tmpfs,omitempty"`
	Cluster     *ServiceVolumeCluster `yaml:",omitempty" json:"cluster,omitempty"`
}

// ServiceVolumeBind are options for a service volume of type bind
type ServiceVolumeBind struct {
	Propagation string `yaml:",omitempty" json:"propagation,omitempty"`
}

// ServiceVolumeVolume are options for a service volume of type volume
type ServiceVolumeVolume struct {
	NoCopy bool `mapstructure:"nocopy" yaml:"nocopy,omitempty" json:"nocopy,omitempty"`
}

// ServiceVolumeTmpfs are options for a service volume of type tmpfs
type ServiceVolumeTmpfs struct {
	Size int64 `yaml:",omitempty" json:"size,omitempty"`
}

// ServiceVolumeCluster are options for a service volume of type cluster.
// Deliberately left blank for future options, but unused now.
type ServiceVolumeCluster struct{}
