package osutility

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Filesystem interface {
	WriteFile(ctx context.Context, filename string, contents []byte, perm fs.FileMode) error
	OpenFileWithPermission(ctx context.Context, filename string, flag int, perm fs.FileMode) (*os.File, error)
	ReadFile(ctx context.Context, filename string) ([]byte, error)
	MkdirAll(ctx context.Context, filename string, perm fs.FileMode) error
	Exists(ctx context.Context, filename string) (bool, error)
	RemoveAll(ctx context.Context, filename string) error
	Remove(ctx context.Context, filename string) error
	Open(ctx context.Context, filename string) (*os.File, error)
	Chown(ctx context.Context, filename string, uid int, gid int) error
	Copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error)
	WriteNewLine(ctx context.Context, filename string, contents []byte) error
	DeleteLineFromFileByKey(ctx context.Context, filename, key string) error
	ExtractTarFile(ctx context.Context, src, dst string) error
	DownloadFileUsingHttp(ctx context.Context, url, filename string, perm fs.FileMode) ([]byte, error)
}

type FakeFilesystem struct{}

func (f *FakeFilesystem) DeleteLineFromFileByKey(ctx context.Context, filename, key string) error {
	logger := log.From(ctx)
	logger.Info("Deleting a line from file", "filename", filename, "Key", key)
	return nil
}

func (f *FakeFilesystem) WriteNewLine(ctx context.Context, filename string, contents []byte) error {
	logger := log.From(ctx)
	logger.Info("Writing new line to file", "filename", filename, "contents", string(contents))
	return nil
}

var _ Filesystem = &FakeFilesystem{}

func NewFakeFilesystem() *FakeFilesystem {
	f := &FakeFilesystem{}

	return f
}

func (f *FakeFilesystem) WriteFile(ctx context.Context, filename string, contents []byte, perm fs.FileMode) error {
	logger := log.From(ctx)
	logger.Info("Writing file", "filename", filename, "contents", string(contents))

	return nil
}

func (f *FakeFilesystem) OpenFileWithPermission(ctx context.Context, filename string, flag int, perm fs.FileMode) (*os.File, error) {
	logger := log.From(ctx)
	logger.Info("Opening file", "filename", filename, "flag", flag, "perm", int(perm))

	return nil, nil
}

func (f *FakeFilesystem) ReadFile(ctx context.Context, filename string) ([]byte, error) {
	logger := log.From(ctx)
	logger.Info("Reading file", "filename", filename)

	return nil, nil
}

func (f *FakeFilesystem) MkdirAll(ctx context.Context, filename string, perm fs.FileMode) error {
	logger := log.From(ctx)
	logger.Info("Making directory", "filename", filename, "perm", int(perm))

	return nil
}

func (f *FakeFilesystem) Exists(ctx context.Context, filename string) (bool, error) {
	logger := log.From(ctx)
	logger.Info("Checking if file exists", "filename", filename)

	return false, nil
}

func (f *FakeFilesystem) RemoveAll(ctx context.Context, filename string) error {
	logger := log.From(ctx)
	logger.Info("Removing file (or dir)", "filename", filename)

	return nil
}

func (f *FakeFilesystem) Remove(ctx context.Context, filename string) error {
	logger := log.From(ctx)
	logger.Info("Removing file (or dir)", "filename", filename)
	return nil
}

func (f *FakeFilesystem) Open(ctx context.Context, filename string) (*os.File, error) {
	logger := log.From(ctx)
	logger.Info("Opening file (or dir)", "filename", filename)
	return nil, nil
}

func (f *FakeFilesystem) Chown(ctx context.Context, filename string, uid int, gid int) error {
	logger := log.From(ctx)
	logger.Info("Changing ownership file (or dir)", "filename", filename, "UID", uid, "GID", gid)
	return nil
}

func (f *FakeFilesystem) Copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	logger := log.From(ctx)
	logger.Info("Copying file", "Source", src, "Destination", dst)
	return 0, nil
}

func (f *FakeFilesystem) ExtractTarFile(ctx context.Context, src, dst string) error {
	return nil
}

func (f *FakeFilesystem) DownloadFileUsingHttp(ctx context.Context, url, filename string, perm fs.FileMode) ([]byte, error) {
	return nil, nil
}

type LiveFilesystem struct{}

var _ Filesystem = &LiveFilesystem{}

func NewLiveFilesystem() *LiveFilesystem {
	l := &LiveFilesystem{}

	return l
}

func (l LiveFilesystem) WriteFile(ctx context.Context, filename string, contents []byte, perm fs.FileMode) error {
	return os.WriteFile(filename, contents, perm)
}

func (l LiveFilesystem) DeleteLineFromFileByKey(ctx context.Context, filename, key string) error {
	file, err := l.Open(ctx, filename)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, key) {
			_, err = buf.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	err = l.WriteFile(ctx, filename, buf.Bytes(), constants.FileReadWriteAccess)
	if err != nil {
		return err
	}
	return nil
}

func (l LiveFilesystem) WriteNewLine(ctx context.Context, filename string, contents []byte) error {
	f, err := l.OpenFileWithPermission(ctx, filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.FileReadWriteAccess)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	if err != nil {
		return err
	}
	var newLine = []byte(string(contents) + "\n")
	if _, err := f.Write(newLine); err != nil {
		return err
	}
	return nil
}

func (l LiveFilesystem) OpenFileWithPermission(ctx context.Context, filename string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(filename, flag, perm)
}

func (l LiveFilesystem) ReadFile(ctx context.Context, filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (l LiveFilesystem) MkdirAll(ctx context.Context, filename string, perm fs.FileMode) error {
	return os.MkdirAll(filename, perm)
}

func (l LiveFilesystem) Exists(ctx context.Context, filename string) (bool, error) {
	logger := log.From(ctx)
	if filename == "" {
		return false, errors.New("filename is blank")
	}
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	logger.Error(err, "error occurred while checking the file", "filename", filename)
	return false, err
}

func (l LiveFilesystem) RemoveAll(ctx context.Context, filename string) error {
	logger := log.From(ctx)

	if filename == "" {
		return errors.New("filename is blank")
	}

	logger.Info("Removing file (or dir)", "filename", filename)

	return os.RemoveAll(filename)
}

func (l LiveFilesystem) Remove(ctx context.Context, filename string) error {
	logger := log.From(ctx)

	if filename == "" {
		return errors.New("filename is blank")
	}

	logger.Info("Removing file (or dir)", "filename", filename)

	return os.Remove(filename)
}

func (l LiveFilesystem) Open(ctx context.Context, filename string) (*os.File, error) {
	logger := log.From(ctx)

	if filename == "" {
		return nil, errors.New("filename is blank")
	}

	logger.Info("Opening file (or dir)", "filename", filename)

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (l LiveFilesystem) Chown(ctx context.Context, filename string, uid int, gid int) error {
	logger := log.From(ctx)

	if filename == "" {
		return errors.New("filename is blank")
	}

	logger.Info("Changing ownership of file (or dir)", "filename", filename, "UID", uid, "GID", gid)

	err := os.Chown(filename, uid, gid)
	if err != nil {
		return err
	}

	return nil
}

func (l LiveFilesystem) Copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	logger := log.From(ctx)

	logger.Info("Copy datatype", "Source", fmt.Sprintf("%T", src), "Destination", fmt.Sprintf("%T", dst))

	bytesCopied, err := io.Copy(dst, src)
	if err != nil {
		return 0, err
	}

	return bytesCopied, nil
}

func (l LiveFilesystem) ExtractTarFile(ctx context.Context, src, dst string) error {
	// Open the tar.gz file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer func(gzr *gzip.Reader) {
		err := gzr.Close()
		if err != nil {
		}
	}(gzr)

	// Create a tar reader
	tarReader := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %v", err)
		}

		// Determine the output path
		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create a directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			// Create a regular file
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				err := outFile.Close()
				if err != nil {
					return err
				}
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			err = outFile.Close()
			if err != nil {
				return err
			}
		default:
			// Handle other file types if necessary
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}
	return nil
}

func (l *LiveFilesystem) DownloadFileUsingHttp(ctx context.Context, url, filename string, perm fs.FileMode) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return nil, nil
}
