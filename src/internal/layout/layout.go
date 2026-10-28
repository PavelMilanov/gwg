// Package layout defines the module-owned filesystem layout of new gwu builds.
// Existing production paths are not automatically discovered or migrated.
package layout

const (
	BaseDir      = "/etc/wireguard/gwu"
	WireGuardDir = BaseDir + "/wireguard"
	QoSDir       = BaseDir + "/qos"
)
