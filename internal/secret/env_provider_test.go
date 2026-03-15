package secret

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvProvider_CanResolve(t *testing.T) {
	provider := NewEnvProvider()

	assert.True(t, provider.CanResolve("env"))
	assert.False(t, provider.CanResolve("keyring"))
	assert.False(t, provider.CanResolve("unknown"))
}

func TestEnvProvider_IsAvailable(t *testing.T) {
	provider := NewEnvProvider()
	assert.True(t, provider.IsAvailable())
}

func TestEnvProvider_Resolve(t *testing.T) {
	provider := NewEnvProvider()
	ctx := context.Background()

	t.Run("existing environment variable", func(t *testing.T) {
		// Set a test environment variable
		key := "TEST_SECRET_VAR"
		value := "test-secret-value"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		ref := Ref{
			Type: "env",
			Name: key,
		}

		result, err := provider.Resolve(ctx, ref)

		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("non-existing environment variable", func(t *testing.T) {
		ref := Ref{
			Type: "env",
			Name: "NON_EXISTING_VAR",
		}

		_, err := provider.Resolve(ctx, ref)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or empty")
	})

	t.Run("empty environment variable", func(t *testing.T) {
		key := "EMPTY_VAR"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		ref := Ref{
			Type: "env",
			Name: key,
		}

		_, err := provider.Resolve(ctx, ref)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or empty")
	})

	t.Run("wrong secret type", func(t *testing.T) {
		ref := Ref{
			Type: "keyring",
			Name: "test",
		}

		_, err := provider.Resolve(ctx, ref)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot resolve secret type")
	})
}

func TestEnvProvider_Store(t *testing.T) {
	provider := NewEnvProvider()
	ctx := context.Background()

	ref := Ref{
		Type: "env",
		Name: "test",
	}

	err := provider.Store(ctx, ref, "value")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support storing")
}

func TestEnvProvider_Delete(t *testing.T) {
	provider := NewEnvProvider()
	ctx := context.Background()

	ref := Ref{
		Type: "env",
		Name: "test",
	}

	err := provider.Delete(ctx, ref)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support deleting")
}

func TestEnvProvider_List(t *testing.T) {
	provider := NewEnvProvider()
	ctx := context.Background()

	// Set some test environment variables
	testVars := map[string]string{
		"TEST_API_KEY":  "sk-1234567890abcdef",
		"TEST_PASSWORD": "secretpassword123",
		"TEST_REGULAR":  "localhost",
		"TEST_SHORT":    "abc",
	}

	// Set environment variables and collect keys for cleanup
	var keysToCleanup []string
	for key, value := range testVars {
		os.Setenv(key, value)
		keysToCleanup = append(keysToCleanup, key)
	}

	// Clean up environment variables after test
	defer func() {
		for _, key := range keysToCleanup {
			os.Unsetenv(key)
		}
	}()

	refs, err := provider.List(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, refs)

	// Should contain our test secret-like variables
	foundSecrets := make(map[string]bool)
	for _, ref := range refs {
		assert.Equal(t, "env", ref.Type)
		if _, exists := testVars[ref.Name]; exists {
			foundSecrets[ref.Name] = true
		}
	}

	// Should detect API key and password as secrets
	assert.True(t, foundSecrets["TEST_API_KEY"] || foundSecrets["TEST_PASSWORD"],
		"Should detect at least one of the secret-like variables")
}

func TestEnvProvider_ResolveWithFallback(t *testing.T) {
	ctx := context.Background()

	// Create a resolver with both env and keyring providers
	resolver := NewResolver()

	// Store a value in keyring
	keyringRef := Ref{
		Type:     "keyring",
		Name:     "TEST_FALLBACK_VAR",
		Original: "${keyring:TEST_FALLBACK_VAR}",
	}
	testValue := "fallback-secret-value"
	err := resolver.Store(ctx, keyringRef, testValue)
	// Skip test if keyring is not available
	if err != nil {
		t.Skip("Keyring not available, skipping fallback test")
		return
	}

	// Get the env provider from the resolver
	envProvider := resolver.providers["env"].(*EnvProvider)

	t.Run("fallback to keyring when env var not set", func(t *testing.T) {
		// Make sure the env var is NOT set
		os.Unsetenv("TEST_FALLBACK_VAR")

		ref := Ref{
			Type:     "env",
			Name:     "TEST_FALLBACK_VAR",
			Original: "${env:TEST_FALLBACK_VAR}",
		}

		result, err := envProvider.Resolve(ctx, ref)

		assert.NoError(t, err)
		assert.Equal(t, testValue, result)
	})

	t.Run("env var takes precedence over keyring", func(t *testing.T) {
		// Set the env var
		envValue := "env-value"
		os.Setenv("TEST_FALLBACK_VAR", envValue)
		defer os.Unsetenv("TEST_FALLBACK_VAR")

		ref := Ref{
			Type:     "env",
			Name:     "TEST_FALLBACK_VAR",
			Original: "${env:TEST_FALLBACK_VAR}",
		}

		result, err := envProvider.Resolve(ctx, ref)

		assert.NoError(t, err)
		assert.Equal(t, envValue, result)
	})
}
