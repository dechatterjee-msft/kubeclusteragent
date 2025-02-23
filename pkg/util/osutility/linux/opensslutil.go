package linux

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"
)

type Openssl interface {
	GenerateCertKeyPair(ctx context.Context, daysUntilExpiry, keyFilePath, certFilePath, subj string) error
	GenerateCSRKeyPair(ctx context.Context, keyFilePath, CSRFilePath, subj string) error
	SignCSR(ctx context.Context, CSRFilePath, daysUntilExpiry, CACertFilePath, CAKeyFilePath, outputCertFilePath, inputExtFilePath string) error
}

type FakeOpenssl struct{}

var _ Openssl = &FakeOpenssl{}

func NewFakeOpenssl() *FakeOpenssl {
	f := &FakeOpenssl{}

	return f
}

func (f *FakeOpenssl) GenerateCertKeyPair(ctx context.Context, daysUntilExpiry, keyFilePath, certFilePath, subj string) error {
	logger := log.From(ctx)
	logger.Info("Generating cert key pair", "daysUntilExpiry", daysUntilExpiry, "keyFilePath", keyFilePath, "certFilePath", certFilePath, "subj", subj)
	return nil
}

func (f *FakeOpenssl) GenerateCSRKeyPair(ctx context.Context, keyFilePath, csrFilePath, subj string) error {
	logger := log.From(ctx)
	logger.Info("Generating csr key pair", "keyFilePath", keyFilePath, "CSRFilePath", csrFilePath, "subj", subj)
	return nil
}

func (f *FakeOpenssl) SignCSR(ctx context.Context, csrFilePath, daysUntilExpiry, caCertFilePath, caKeyFilePath, outputCertFilePath, inputExtFilePath string) error {
	logger := log.From(ctx)
	logger.Info("Signing CSR using generated cert", "CSRFilePath", csrFilePath, "daysUntilExpiry", daysUntilExpiry, "CACertFilePath", caCertFilePath, "CAKeyFilePath", caKeyFilePath, "outputCertFilePath", outputCertFilePath, "inputExtFilePath", inputExtFilePath)
	return nil
}

type LiveOpenssl struct {
	exec Exec
}

var _ Openssl = &LiveOpenssl{}

func NewLiveOpenssl(execUtil Exec) *LiveOpenssl {
	l := &LiveOpenssl{
		exec: execUtil,
	}
	return l
}

func (l *LiveOpenssl) GenerateCertKeyPair(ctx context.Context, daysUntilExpiry, keyFilePath, certFilePath, subj string) error {
	logger := log.From(ctx)
	logger.Info("Generating cert key pair", "daysUntilExpiry", daysUntilExpiry, "keyFilePath", keyFilePath, "certFilePath", certFilePath, "subj", subj)
	_, _, err := l.exec.Command(ctx, "openssl", nil, []string{"req", "-x509", "-sha256", "-newkey", "rsa:4096", "-days", daysUntilExpiry, "-nodes", "-keyout", keyFilePath, "-out", certFilePath, "-subj", subj}...)
	return err
}

func (l *LiveOpenssl) GenerateCSRKeyPair(ctx context.Context, keyFilePath, csrFilePath, subj string) error {
	logger := log.From(ctx)
	logger.Info("Generating csr key pair", "keyFilePath", keyFilePath, "CSRFilePath", csrFilePath, "subj", subj)
	_, _, err := l.exec.Command(ctx, "openssl", nil, []string{"req", "-sha256", "-newkey", "rsa:4096", "-nodes", "-keyout", keyFilePath, "-out", csrFilePath, "-subj", subj}...)
	return err
}

func (l *LiveOpenssl) SignCSR(ctx context.Context, csrFilePath, daysUntilExpiry, caCertFilePath, caKeyFilePath, outputCertFilePath, inputExtFilePath string) error {
	logger := log.From(ctx)
	logger.Info("Signing CSR using generated cert", "CSRFilePath", csrFilePath, "daysUntilExpiry", daysUntilExpiry, "CACertFilePath", caCertFilePath, "CAKeyFilePath", caKeyFilePath, "outputCertFilePath", outputCertFilePath, "inputExtFilePath", inputExtFilePath)
	_, _, err := l.exec.Command(ctx, "openssl", nil, []string{"x509", "-req", "-sha256", "-in", csrFilePath, "-days", daysUntilExpiry, "-CA", caCertFilePath, "-CAkey", caKeyFilePath, "-CAcreateserial", "-out", outputCertFilePath, "-extfile", inputExtFilePath}...)
	return err
}
