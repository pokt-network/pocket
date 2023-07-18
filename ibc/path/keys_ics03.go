package path

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// ICS03
// The following paths are the keys to the store as defined in:
// https://github.com/cosmos/ibc/blob/master/spec/core/ics-003-connection-semantics#store-paths
////////////////////////////////////////////////////////////////////////////////

// clientConnectionsPath defines a reverse mapping from clients to a set of connections
func clientConnectionsPath(clientID string) string {
	return fullClientPath(clientID, KeyConnectionPrefix)
}

// ClientConnectionsKey returns the store key for the connections of a given client
func ClientConnectionsKey(clientID string) []byte {
	return []byte(clientConnectionsPath(clientID))
}

// ClientConnectionPath defines the path under which the connections are stored in the client store
func ClientConnectionPath(clientID string) string {
	return clientPath(clientID, KeyConnectionPrefix)
}

// connectionPath defines the path under which connection paths are stored
func connectionPath(connectionID string) string {
	return fmt.Sprintf("%s/%s", KeyConnectionPrefix, connectionID)
}

// ConnectionKey returns the store key for a particular connection
func ConnectionKey(connectionID string) []byte {
	return []byte(connectionPath(connectionID))
}

// ConnectionPath returns the path for a particular connection in the connections store
func ConnectionPath(connectionID string) string {
	return connectionID
}
