package path

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// ICS05
// The following paths are the keys to the store as defined in
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-005-port-allocation#store-paths
////////////////////////////////////////////////////////////////////////////////

// portPath defines the path under which ports paths are stored on the capability module
func portPath(portID string) string {
	return fmt.Sprintf("%s/%s", KeyPortPrefix, portID)
}

// PortKey returns the store key for a port in the capability module
func PortKey(portID string) []byte {
	return []byte(portPath(portID))
}

// PortPath returns the path under which the Port is stored in the Port store
func PortPath(portID string) string {
	return portID
}
