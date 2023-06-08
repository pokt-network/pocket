package flags

var (
	// RemoveCLIURL is the URL of the remote RPC node which the CLI will interact with.
	// Formatted as <protocol>://<host> (uses RPC Port).
	// (see: --help the root command for more info).
	RemoteCLIURL string

	// DataDir a path to store pocket related data (keybase etc.).
	// (see: --help the root command for more info).
	DataDir string

	// ConfigPath is the path to the node config file.
	// (see: --help the root command for more info).
	ConfigPath string

	// If true skips the interactive prompts wherever possible (useful for scripting & automation)
	// (see: --help the root command for more info).
	NonInteractive bool

	// Show verbose output
	// (see: --help the root command for more info).
	Verbose bool
)
