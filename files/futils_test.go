package files

import "testing"

func TestNormalizationPath(t *testing.T) {
	p := "C:\\Users\\gopher\\Documents"

	normalized := NormalizePath(p)
	if normalized != "/c/Users/gopher/Documents" {
		t.Fatal("path normalization failed")
	}

	p2 := UnNormalizePath(normalized)
	if p2 != p {
		t.Fatal("path un-normalization failed")
	}
}
