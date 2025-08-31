package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Create a temporary .env file for testing
	content := []byte("TEST_KEY=test_value\n")
	err := os.WriteFile(".env", content, 0644)
	assert.NoError(t, err, "Failed to create test .env file")
	defer func() { _ = os.Remove(".env") }()

	// Test Load function
	err = Load()
	assert.NoError(t, err, "Load() should not return error")

	// Verify environment variable is loaded correctly
	expected := "test_value"
	actual := os.Getenv("TEST_KEY")
	assert.Equal(t, expected, actual, "Environment variable should be loaded correctly")
}

func TestGet(t *testing.T) {
	// Set up test environment variable
	err := os.Setenv("TEST_ENV_KEY", "test_env_value")
	assert.NoError(t, err)
	defer func() { _ = os.Unsetenv("TEST_ENV_KEY") }()

	// Test Get function
	expected := "test_env_value"
	actual := Get("TEST_ENV_KEY")
	assert.Equal(t, expected, actual, "Should return correct environment variable value")

	// Test non-existent environment variable
	nonExistent := Get("NON_EXISTENT_KEY")
	assert.Empty(t, nonExistent, "Should return empty string for non-existent key")
}
