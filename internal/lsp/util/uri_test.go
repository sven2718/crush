package util

import (
	"runtime"
	"testing"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/stretchr/testify/require"
)

// TestURIToPath_RoundTrip pins the Windows behavior: powernap's Path()
// returns a spurious leading backslash before the drive letter; URIToPath
// must strip it.
func TestURIToPath_RoundTrip(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("drive-letter path round-trip is Windows-specific")
	}

	cases := []struct {
		name     string
		in       string
		wantPath string
	}{
		{
			name:     "simple drive path",
			in:       `C:\dev\Leviathan\foo.lua`,
			wantPath: `C:\dev\Leviathan\foo.lua`,
		},
		{
			name:     "path with spaces",
			in:       `C:\dev\Leviathan\resources\Lua state\foo.lua`,
			wantPath: `C:\dev\Leviathan\resources\Lua state\foo.lua`,
		},
		{
			name:     "forward-slash input normalizes to backslashes",
			in:       `C:/dev/Leviathan/foo.lua`,
			wantPath: `C:\dev\Leviathan\foo.lua`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			uri := protocol.URIFromPath(tc.in)
			got, err := URIToPath(uri)
			require.NoError(t, err)
			require.Equal(t, tc.wantPath, got)
		})
	}
}

// TestNormalizeFilePath covers the pure-function normalizer without going
// through the LSP URI machinery.
func TestNormalizeFilePath(t *testing.T) {
	t.Parallel()

	// Inputs that must normalize identically on every platform.
	sharedCases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
	}
	for _, tc := range sharedCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, NormalizeFilePath(tc.in))
		})
	}

	if runtime.GOOS != "windows" {
		return
	}

	t.Run("strips spurious leading backslash", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, `C:\dev\foo`, NormalizeFilePath(`\C:\dev\foo`))
	})
	t.Run("strips spurious leading forward slash", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, `C:\dev\foo`, NormalizeFilePath(`/C:/dev/foo`))
	})
	t.Run("leaves real drive paths alone", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, `C:\dev\foo`, NormalizeFilePath(`C:\dev\foo`))
	})
	t.Run("leaves UNC paths alone", func(t *testing.T) {
		t.Parallel()
		// filepath.Clean will collapse `\\` to `\\` on UNC heads.
		require.Equal(t, `\\server\share\foo`, NormalizeFilePath(`\\server\share\foo`))
	})
}

// TestSamePath checks that comparisons survive drive-letter case drift and
// separator drift, which are the two ways we have seen LSP paths diverge
// from tool-argument paths on Windows.
func TestSamePath(t *testing.T) {
	t.Parallel()

	t.Run("identical paths compare equal", func(t *testing.T) {
		t.Parallel()
		require.True(t, SamePath(`/tmp/foo`, `/tmp/foo`))
	})

	if runtime.GOOS != "windows" {
		return
	}

	t.Run("drive letter case drift", func(t *testing.T) {
		t.Parallel()
		require.True(t, SamePath(`C:\dev\foo`, `c:\dev\foo`))
	})
	t.Run("separator drift", func(t *testing.T) {
		t.Parallel()
		require.True(t, SamePath(`C:\dev\foo`, `C:/dev/foo`))
	})
	t.Run("spurious leading separator on one side", func(t *testing.T) {
		t.Parallel()
		require.True(t, SamePath(`\C:\dev\foo`, `C:\dev\foo`))
	})
	t.Run("distinct paths compare unequal", func(t *testing.T) {
		t.Parallel()
		require.False(t, SamePath(`C:\dev\foo`, `C:\dev\bar`))
	})
}
