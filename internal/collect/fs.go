package collect

import (
	"os"
	"path/filepath"
	"strings"
)

// joinFS joins fsRoot with path elements.
func joinFS(fsRoot string, elems ...string) string {
	if fsRoot == "" {
		return filepath.Join(elems...)
	}
	return filepath.Join(append([]string{fsRoot}, elems...)...)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func isDotFile(name string) bool {
	return strings.Contains(name, ".")
}
