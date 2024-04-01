package util

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func JsonArray(values []string) []byte {
	var sb bytes.Buffer
	sb.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Quote(value))
	}
	sb.WriteByte(']')
	return sb.Bytes()
}

func CloneRepo(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		// Already exists
		// TODO: ensure up-to-date?
		return nil
	}

	const repo = "https://github.com/KhronosGroup/OpenGL-Refpages.git"
	git := exec.Command("git", "clone", repo, path)
	git.Stdout = os.Stdout
	git.Stderr = os.Stderr
	return git.Run()
}

func RepoPath() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); len(xdg) > 0 {
		return filepath.Join(xdg, "gldoc")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "gldoc")
}

func Sanitize[T ~string](text T) string {
	var sb strings.Builder
	for _, line := range strings.Split(string(text), "\n") {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strings.TrimSpace(line))
	}

	return sb.String()
}

func DirNames(gles bool, version float64) (names []string) {
	if gles {
		switch {
		case version == 0 || version > 3.1:
			names = append(names, "es3")
			fallthrough
		case version > 3.0:
			names = append(names, "es3.1")
			fallthrough
		case version > 2.0:
			names = append(names, "es3.0")
			fallthrough
		case version > 1.0:
			names = append(names, "es2.0")
			fallthrough
		default:
			names = append(names, "es1.0")
		}
	} else {
		if version == 0 || version > 2.1 {
			names = append(names, "gl4")
		}
		names = append(names, "gl2.1")
	}

	return
}
