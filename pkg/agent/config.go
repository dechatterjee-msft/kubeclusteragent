package agent

// Config is configuration for the automation app.
type Config struct {
	// GRPCAddr is the address of the GRPC server.
	GRPCAddr string

	// ServerAddr is the address of the HTTP server.
	ServerAddr string

	// DryRun runs the application in dry-run mode.
	DryRun bool

	// StateFilePath is where the state file is stored.
	StateFilePath string

	// TokenSharedKey is the shared key for verifying tokens.
	TokenSharedKey string

	// ServerKeyFilePath points to the server key used for tls.
	ServerKeyFilePath string

	// ServerCertFilePath points to the server cert used for tls.
	ServerCertFilePath string

	// CACertFilePath points to the ca cert used for tls in grpc gateway.
	CACertFilePath string
	// PrimaryNetwork Interface
	PrimaryNetworkInterface string
}
