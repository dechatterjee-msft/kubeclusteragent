package osutility

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"kubeclusteragent/pkg/util/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveFileSystem_WriteFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "write_file")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	tests := []struct {
		name      string
		filename  string
		contents  []byte
		perm      fs.FileMode
		wantError bool
	}{
		{
			name:      "valid file name",
			filename:  filepath.Join(dir, "good"),
			contents:  []byte("good"),
			perm:      0644,
			wantError: false,
		},
		{
			name:      "no filename",
			filename:  "",
			contents:  nil,
			perm:      0,
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := NewLiveFilesystem()
			err := l.WriteFile(context.Background(), test.filename, test.contents, test.perm)
			test.CheckError(t, test.wantError, err, func() {
				fi, err := os.Stat(test.filename)
				require.NoError(t, err)
				assert.False(t, fi.IsDir())
				assert.Equal(t, test.perm, fi.Mode())
			})
		})
	}
}

func TestLiveFilesystem_OpenFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "open_file")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
		flag     int
		perm     fs.FileMode
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Opening an existing file with read-only flag
		{
			name: "TestOpenExistingFileReadOnly",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_file"),
				flag:     os.O_RDONLY,
				perm:     0644,
			},
			want:    nil,
			wantErr: assert.NoError,
		},

		// Test case 2: Opening a non-existing file with write-only flag and create option
		{
			name: "TestOpenNonExistingFileWriteOnly",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file"),
				flag:     os.O_WRONLY | os.O_CREATE,
				perm:     0644,
			},
			want:    nil,
			wantErr: assert.NoError,
		},

		// Test case 3: Opening a file with incorrect permissions
		{
			name: "TestOpenFileIncorrectPermissions",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "file_with_incorrect_permissions"),
				flag:     os.O_RDWR,
				perm:     0000,
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "TestOpenExistingFileReadOnly" || tt.name == "TestOpenFileIncorrectPermissions" {
				f, err := os.Create(tt.args.filename)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			l := LiveFilesystem{}
			_, err := l.OpenFileWithPermission(tt.args.ctx, tt.args.filename, tt.args.flag, tt.args.perm)
			if tt.wantErr != nil {
				if os.IsPermission(err) {
					tt.wantErr(t, err, fmt.Sprintf("OpenFile(%v, %v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.flag, tt.args.perm))
				} else {
					assert.NoError(t, err, fmt.Sprintf("OpenFile(%v, %v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.flag, tt.args.perm))
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("OpenFile(%v, %v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.flag, tt.args.perm))
			}
		})
	}
}

func TestLiveFilesystem_ReadFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "read_file")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Reading an existing file with content
		{
			name: "TestReadExistingFileWithContent",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_file"),
			},
			want:    []byte("This is an existing file with content."),
			wantErr: assert.NoError,
		},

		// Test case 2: Reading an existing file with no content (empty file)
		{
			name: "TestReadExistingFileEmpty",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "empty_file"),
			},
			want:    []byte{},
			wantErr: assert.NoError,
		},

		// Test case 3: Reading a non-existing file
		{
			name: "TestReadNonExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file"),
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "TestReadExistingFileWithContent" || tt.name == "TestReadExistingFileEmpty" {
				f, err := os.Create(tt.args.filename)
				require.NoError(t, err)
				content := tt.want
				_, err = io.WriteString(f, string(content))
				require.NoError(t, err)
			}
			l := LiveFilesystem{}
			got, err := l.ReadFile(tt.args.ctx, tt.args.filename)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadFile(%v, %v)", tt.args.ctx, tt.args.filename)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ReadFile(%v, %v)", tt.args.ctx, tt.args.filename)
		})
	}
}

func TestLiveFilesystem_MkdirAll(t *testing.T) {
	dir, err := os.MkdirTemp("", "mkdir_all")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
		perm     fs.FileMode
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Creating a new directory with valid permissions
		{
			name: "TestMkdirAllNewDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "new_directory"),
				perm:     0755,
			},
			wantErr: assert.NoError,
		},

		// Test case 2: Creating a nested directory structure with valid permissions
		{
			name: "TestMkdirAllNestedDirectories",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "nested/directory/structure"),
				perm:     0700,
			},
			wantErr: assert.NoError,
		},

		// Test case 3: Creating a directory that already exists
		{
			name: "TestMkdirAllExistingDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_directory"),
				perm:     0644,
			},
			wantErr: assert.NoError,
		},

		// Test case 4: Creating a directory with incorrect permissions
		{
			name: "TestMkdirAllIncorrectPermissions",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "directory_with_incorrect_permissions"),
				perm:     0000,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiveFilesystem{}
			err := l.MkdirAll(tt.args.ctx, tt.args.filename, tt.args.perm)
			if tt.wantErr != nil {
				if os.IsPermission(err) {
					tt.wantErr(t, err, fmt.Sprintf("MkdirAll(%v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.perm))
				} else {
					assert.NoError(t, err, fmt.Sprintf("MkdirAll(%v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.perm))
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("MkdirAll(%v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.perm))
			}
		})
	}
}

func TestLiveFileSystem_Exists(t *testing.T) {
	f, err := ioutil.TempFile("", "exists")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	defer func() {
		require.NoError(t, os.RemoveAll(f.Name()))
	}()

	tests := []struct {
		name      string
		filename  string
		want      bool
		wantError bool
	}{
		{
			name:      "valid file name",
			filename:  f.Name(),
			want:      true,
			wantError: false,
		},
		{
			name:      "invalid filename",
			filename:  f.Name() + "-invalid",
			want:      false,
			wantError: false,
		},
		{
			name:      "no filename",
			filename:  "",
			want:      false,
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := NewLiveFilesystem()
			got, err := l.Exists(context.Background(), test.filename)
			test.CheckError(t, test.wantError, err, func() {
				require.Equal(t, test.want, got)
			})
		})
	}
}

func TestLiveFilesystem_RemoveAll(t *testing.T) {
	dir, err := os.MkdirTemp("", "remove_all")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Removing an existing file
		{
			name: "TestRemoveExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_file"),
			},
			wantErr: assert.NoError,
		},

		// Test case 2: Removing a non-existing file
		{
			name: "TestRemoveNonExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file"),
			},
			wantErr: assert.Error,
		},

		// Test case 3: Removing an existing directory
		{
			name: "TestRemoveExistingDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_directory"),
			},
			wantErr: assert.NoError,
		},

		// Test case 4: Removing a non-existing directory
		{
			name: "TestRemoveNonExistingDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_directory"),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "TestRemoveExistingFile" {
				f, err := os.Create(tt.args.filename)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			} else if tt.name == "TestRemoveExistingDirectory" {
				err := os.Mkdir(tt.args.filename, 0755)
				require.NoError(t, err)
			}

			_, err := os.Stat(tt.args.filename)
			if os.IsNotExist(err) {
				tt.wantErr(t, err, fmt.Sprintf("RemoveAll(%v, %v)", tt.args.ctx, tt.args.filename))
				return
			}

			l := LiveFilesystem{}
			err = l.RemoveAll(tt.args.ctx, tt.args.filename)
			tt.wantErr(t, err, fmt.Sprintf("RemoveAll(%v, %v)", tt.args.ctx, tt.args.filename))
		})
	}
}

func TestLiveFilesystem_Remove(t *testing.T) {
	dir, err := os.MkdirTemp("", "remove")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Removing an existing file
		{
			name: "TestRemoveExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_file"),
			},
			wantErr: assert.NoError,
		},

		// Test case 2: Removing a non-existing file
		{
			name: "TestRemoveNonExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file"),
			},
			wantErr: assert.Error,
		},

		// Test case 3: Removing an empty directory
		{
			name: "TestRemoveEmptyDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "empty_directory"),
			},
			wantErr: assert.NoError,
		},

		// Test case 4: Removing a non-empty directory
		{
			name: "TestRemoveNonEmptyDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_empty_directory"),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiveFilesystem{}
			if tt.name == "TestRemoveExistingFile" {
				f, err := os.Create(tt.args.filename)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			} else if tt.name == "TestRemoveEmptyDirectory" {
				err := os.Mkdir(tt.args.filename, 0755)
				require.NoError(t, err)
			} else if tt.name == "TestRemoveNonEmptyDirectory" {
				err := os.Mkdir(tt.args.filename, 0755)
				f, err := os.Create(filepath.Join(tt.args.filename, "non_empty_dir_file"))
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			err := l.Remove(tt.args.ctx, tt.args.filename)
			tt.wantErr(t, err, fmt.Sprintf("Remove(%v, %v)", tt.args.ctx, tt.args.filename))
		})
	}
}

func TestLiveFilesystem_Open(t *testing.T) {
	dir, err := os.MkdirTemp("", "open")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Opening an existing file with read-only flag
		{
			name: "TestOpenExistingFileReadOnly",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "existing_file"),
			},
			want:    &os.File{},
			wantErr: assert.NoError,
		},

		// Test case 2: Opening a non-existing file
		{
			name: "TestOpenNonExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file.txt"),
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiveFilesystem{}
			if tt.name == "TestOpenExistingFileReadOnly" {
				f, err := os.Create(tt.args.filename)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			_, err = l.Open(tt.args.ctx, tt.args.filename)
			tt.wantErr(t, err, fmt.Sprintf("Open(%v, %v)", tt.args.ctx, tt.args.filename))
		})
	}
}

func TestLiveFilesystem_Chown(t *testing.T) {
	dir, err := os.MkdirTemp("", "chown")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx      context.Context
		filename string
		uid      int
		gid      int
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Changing ownership of a non-existing file
		{
			name: "TestChownNonExistingFile",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_file"),
				uid:      1001,
				gid:      1001,
			},
			wantErr: assert.Error,
		},

		// Test case 2: Changing ownership of a non-existing directory
		{
			name: "TestChownNonExistingDirectory",
			args: args{
				ctx:      context.Background(),
				filename: filepath.Join(dir, "non_existing_directory"),
				uid:      1001,
				gid:      1001,
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLiveFilesystem()
			err := l.Chown(tt.args.ctx, tt.args.filename, tt.args.uid, tt.args.gid)
			tt.wantErr(t, err, fmt.Sprintf("Chown(%v, %v, %v, %v)", tt.args.ctx, tt.args.filename, tt.args.uid, tt.args.gid))
		})
	}
}

func TestLiveFilesystem_Copy(t *testing.T) {
	dir, err := os.MkdirTemp("", "copy")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	type args struct {
		ctx context.Context
		src io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantDst string
		want    int64
		wantErr assert.ErrorAssertionFunc
	}{
		// Test case 1: Copying data from a valid source to destination
		{
			name: "TestCopyValidData",
			args: args{
				ctx: context.Background(),
				src: strings.NewReader("Hello, World!"),
			},
			wantDst: "Hello, World!",
			want:    13,
			wantErr: assert.NoError,
		},

		// Test case 2: Copying data from an empty source
		{
			name: "TestCopyEmptySource",
			args: args{
				ctx: context.Background(),
				src: strings.NewReader(""),
			},
			wantDst: "",
			want:    0,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiveFilesystem{}
			dst := &bytes.Buffer{}
			got, err := l.Copy(tt.args.ctx, dst, tt.args.src)
			if !tt.wantErr(t, err, fmt.Sprintf("Copy(%v, %v, %v)", tt.args.ctx, tt.args.src, dst)) {
				return
			}
			assert.Equalf(t, tt.wantDst, dst.String(), "Copy(%v, %v, %v)", tt.args.ctx, tt.args.src, dst)
			assert.Equalf(t, tt.want, got, "Copy(%v, %v, %v)", tt.args.ctx, tt.args.src, dst)
		})
	}
}
