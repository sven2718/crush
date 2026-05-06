// Package util provides helpers for bridging LSP protocol types with Go's
// path conventions.
//
// This file exists primarily to paper over a Windows-only quirk in
// powernap's `DocumentURI.Path()`: for a well-formed URI like
// `file:///C:/dev/foo`, Path() returns `\C:\dev\foo` with a spurious
// leading backslash. The cause is in `powernap/pkg/lsp/protocol/uri.go`'s
// `filename()`:
//
//	u, _ := url.ParseRequestURI(uri)    // u.Path == "/C:/dev/foo"
//	if isWindowsDrivePath(u.Path) { ... // FALSE: VolumeName("/C:/...") == ""
//	return u.Path                        // "/C:/dev/foo" → FromSlash → "\C:\dev\foo"
//
// The check should strip the leading `/` before asking about the drive
// letter. Until that is fixed upstream we normalize on the Crush side.
package util

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
)

// URIToPath converts an LSP DocumentURI into a local filesystem path, with
// Windows normalization applied. Prefer this over calling `.Path()` directly
// so callers do not have to repeat the workaround.
func URIToPath(uri protocol.DocumentURI) (string, error) {
	raw, err := uri.Path()
	if err != nil {
		return "", err
	}
	return NormalizeFilePath(raw), nil
}

// NormalizeFilePath canonicalizes `p` for intra-Crush use. On Windows it
// strips a spurious leading separator sitting in front of a drive letter
// (the powernap bug above) and then runs `filepath.Clean`. On POSIX it is
// just `filepath.Clean`.
func NormalizeFilePath(p string) string {
	if p == "" {
		return p
	}
	if runtime.GOOS == "windows" && hasSpuriousLeadingSlash(p) {
		p = p[1:]
	}
	return filepath.Clean(p)
}

// SamePath reports whether `a` and `b` refer to the same filesystem
// location. On Windows comparison is case-insensitive (drive letters and
// component names both); on POSIX it is a byte-exact compare after
// normalization.
//
// Use this instead of `a == b` whenever one side comes from an LSP URI and
// the other from a tool argument - the two sources historically disagree
// on drive-letter casing and separator direction.
func SamePath(a, b string) bool {
	a = NormalizeFilePath(a)
	b = NormalizeFilePath(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// hasSpuriousLeadingSlash returns true when `p` looks like `\C:\...` or
// `/C:/...` - a Windows drive path with one extra separator bolted on
// in front. Only meaningful on Windows.
func hasSpuriousLeadingSlash(p string) bool {
	if len(p) < 3 {
		return false
	}
	if p[0] != '\\' && p[0] != '/' {
		return false
	}
	if !isASCIILetter(p[1]) {
		return false
	}
	return p[2] == ':'
}

func isASCIILetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
