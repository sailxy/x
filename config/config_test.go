package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const content = `
name: hello
age: 10
`

func createTempFile() (string, error) {
	file, err := os.Create("/tmp/config.yaml")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(content)
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestLoad(t *testing.T) {
	filename, err := createTempFile()
	assert.NoError(t, err)
	t.Log(filename)
	defer func() { _ = os.Remove(filename) }()

	data := struct {
		Name string
		Age  int
	}{}

	cfg := New()
	err = cfg.LoadFromFile(filename, &data)
	assert.NoError(t, err)
	t.Log(data)
}
