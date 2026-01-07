package helpers

import (
	"strings"

	"golang.org/x/sys/windows"
)

func CString(s string) (*byte, error) {
	// Windows API expects null-terminated ANSI here (SimConnect uses LPCSTR).
	b, err := windows.BytePtrFromString(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func TrimCString(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		return string(b[:i])
	}
	// safer: find zero manually
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

func HzToMHz(hz int) float64 {
	return float64(hz) / 1_000_000.0
}
