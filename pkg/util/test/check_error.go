package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// CheckError or not is a helper that requires an error or not. If functions are
// present, execute them if there was no error.
func CheckError(t *testing.T, wantErr bool, err error, fns ...func()) {
	if wantErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)

	for _, fn := range fns {
		fn()
	}
}
