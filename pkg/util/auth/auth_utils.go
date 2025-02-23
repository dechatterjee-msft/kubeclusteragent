package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"net/http"
	"os"

	"google.golang.org/grpc/credentials"
)

var logger = log.From(context.Background())

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

func LoadTLSCredentials(serverCertFilePath, serverKeyFilePath string) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(serverCertFilePath, serverKeyFilePath)
	if err != nil {
		logger.Error(err, "error loading cert key pair")
		return nil, err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	return credentials.NewTLS(config), nil
}

func LoadTLSCredentialsForGateway(caCertFilePath string) (credentials.TransportCredentials, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if caCertFilePath != "" {
		caPem, err := os.ReadFile(caCertFilePath)
		if err != nil {
			logger.Error(err, "error reading CA cert")
			return nil, err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPem) {
			err := fmt.Errorf("failed to add CA's certificate")
			logger.Error(err, "")
			return nil, err
		}
		config.RootCAs = certPool
	}
	return credentials.NewTLS(config), nil
}

func AuthWrapperHandler(tokenSharedKey string, handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		jwtManager := CreateJwtManager(tokenSharedKey)
		userClaims, err := jwtManager.VerifyToken(accessToken)
		if err != nil {
			logger.Error(err, "failed to verify access token")
			w.Header().Set("Content-Type", "application/json; charset=utf-8") //TODO: This seems to be overwritten, need to fix
			http.Error(w, getErrorResponse(fmt.Sprintf("failed to verify access token: %s", err.Error())), http.StatusUnauthorized)
			return
		}
		if userClaims.Role != adminKey {
			logger.Info(fmt.Sprintf("%s does not have access to metrics server", userClaims.Username))
			w.Header().Set("Content-Type", "application/json; charset=utf-8") //TODO: This seems to be overwritten, need to fix
			http.Error(w, getErrorResponse("access denied. Metrics server is restricted to admin role"), http.StatusForbidden)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func getErrorResponse(msg string) string {
	errorResponse := &ErrorResponse{
		Message: msg,
	}
	errorString, err := json.Marshal(errorResponse)
	if err != nil {
		return msg // returning plain text in case of error, this scenario will not happen
	}
	return string(errorString)
}

func GenerateCerts(ctx context.Context) error {
	exec := linux.NewLiveExec()

	// generating ext.cnf file with extensions
	err := linux.NewLiveFilesystem().WriteFile(ctx, generatedServerExtFilePath, []byte(serverExtFileContents), 0644)
	if err != nil {
		return fmt.Errorf("error creating ext.cnf file: %w", err)
	}

	openssl := linux.NewLiveOpenssl(exec)
	// generating CAs private key and self signed certificate
	err = openssl.GenerateCertKeyPair(ctx, defaultDaysUntilExpiry, generatedCAKeyFilePath, GeneratedCACertFilePath, certSubjContents)
	if err != nil {
		return err
	}

	// generating web server's private key and certificate signing request
	err = openssl.GenerateCSRKeyPair(ctx, GeneratedServerKeyFilePath, generatedServerCSRFilePath, certSubjContents)
	if err != nil {
		return err
	}

	// get signed certificate by signing server's CSR with CA's private key
	err = openssl.SignCSR(ctx, generatedServerCSRFilePath, defaultDaysUntilExpiry, GeneratedCACertFilePath, generatedCAKeyFilePath, GeneratedServerCertFilePath, generatedServerExtFilePath)
	if err != nil {
		return err
	}

	return nil
}
