package auth

import "kubeclusteragent/pkg/constants"

const (
	GeneratedServerKeyFilePath  = constants.CertsDirectory + "/server-key.pem"
	GeneratedServerCertFilePath = constants.CertsDirectory + "/server-cert.pem"
	generatedCAKeyFilePath      = constants.CertsDirectory + "/ca-key.pem"
	GeneratedCACertFilePath     = constants.CertsDirectory + "/ca-cert.pem"
	generatedServerExtFilePath  = constants.CertsDirectory + "/server-ext.cnf"
	generatedServerCSRFilePath  = constants.CertsDirectory + "/server-req.pem"
	serverExtFileContents       = "subjectAltName=DNS:*.example.com,IP:0.0.0.0"
	certSubjContents            = "/O=example/OU=sebu/CN=*.example.com"
	defaultDaysUntilExpiry      = "365"
	adminKey                    = "admin"
)
