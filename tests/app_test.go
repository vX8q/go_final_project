package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getURL(path string) string {
	path = strings.TrimPrefix(strings.ReplaceAll(path, `\`, `/`), `../web/`)
	return fmt.Sprintf("http://localhost:%d/%s", Port, path)
}

func getBody(path string) ([]byte, error) {
	resp, err := http.Get(getURL(path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func walkDir(path string, f func(fname string) error) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		fname := filepath.Join(path, e.Name())
		if e.IsDir() {
			if err := walkDir(fname, f); err != nil {
				return err
			}
		} else if err := f(fname); err != nil {
			return err
		}
	}
	return nil
}

func TestApp(t *testing.T) {
	cmp := func(fname string) error {
		fbody, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		body, err := getBody(fname)
		if err != nil {
			return err
		}
		assert.Equal(t, len(fbody), len(body), "сервер возвращает для %s данные другого размера", fname)
		return nil
	}
	assert.NoError(t, walkDir("../web", cmp))
}
