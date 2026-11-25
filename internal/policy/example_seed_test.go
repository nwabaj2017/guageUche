package policy

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestShippedPoliciesContainNoSeed(t *testing.T) {
	t.Parallel()

	paths := []string{"../../gauge.example.toml"}
	err := filepath.WalkDir("../../testdata/policy", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata/policy: %v", err)
	}
	if len(paths) < 2 {
		t.Fatalf("found %d shipped policy files, want the example and at least one policy fixture", len(paths))
	}
	sort.Strings(paths)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if containsSeed(data) {
			t.Errorf("%s contains a seed-shaped token", path)
		}
		if hasSeedShape(data) {
			t.Errorf("%s contains a 56-character base32 token starting with S", path)
		}
	}
}

func hasSeedShape(data []byte) bool {
	for i := 0; i < len(data); {
		if !base32Byte(data[i]) {
			i++
			continue
		}
		start := i
		for i < len(data) && base32Byte(data[i]) {
			i++
		}
		if i-start == 56 && (data[start] == 'S' || data[start] == 's') {
			return true
		}
	}
	return false
}
