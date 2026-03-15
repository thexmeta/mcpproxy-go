package secret

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	SecretTypeEnv = "env"
)

// EnvProvider resolves secrets from environment variables
type EnvProvider struct {
	// fallbackResolver is used to resolve env vars from keyring when not found in process env
	fallbackResolver *Resolver
}

// NewEnvProvider creates a new environment variable provider
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// SetFallbackResolver sets a resolver to use for falling back to keyring lookup
// This allows ${env:NAME} to be resolved from keyring if the env var is not set
func (p *EnvProvider) SetFallbackResolver(resolver *Resolver) {
	p.fallbackResolver = resolver
}

// CanResolve returns true if this provider can handle the given secret type
func (p *EnvProvider) CanResolve(secretType string) bool {
	return secretType == SecretTypeEnv
}

// Resolve retrieves the secret value from environment variables.
// If the environment variable is not found and a fallback resolver is configured,
// it will attempt to resolve from keyring using the same name.
func (p *EnvProvider) Resolve(ctx context.Context, ref Ref) (string, error) {
	if !p.CanResolve(ref.Type) {
		return "", fmt.Errorf("env provider cannot resolve secret type: %s", ref.Type)
	}

	// First, try to get from actual environment variables
	value := os.Getenv(ref.Name)
	if value != "" {
		return value, nil
	}

	// If not found in environment and fallback resolver is available,
	// try to resolve from keyring (for env vars set via UI)
	if p.fallbackResolver != nil {
		keyringRef := Ref{
			Type:     "keyring",
			Name:     ref.Name,
			Original: fmt.Sprintf("${keyring:%s}", ref.Name),
		}

		// Try to resolve from keyring
		keyringProvider := p.fallbackResolver.providers["keyring"]
		if keyringProvider != nil && keyringProvider.IsAvailable() {
			keyringValue, err := keyringProvider.Resolve(ctx, keyringRef)
			if err == nil && keyringValue != "" {
				return keyringValue, nil
			}
		}
	}

	return "", fmt.Errorf("environment variable %s not found or empty", ref.Name)
}

// Store is not supported for environment variables
func (p *EnvProvider) Store(_ context.Context, _ Ref, _ string) error {
	return fmt.Errorf("env provider does not support storing secrets")
}

// Delete is not supported for environment variables
func (p *EnvProvider) Delete(_ context.Context, _ Ref) error {
	return fmt.Errorf("env provider does not support deleting secrets")
}

// List returns all environment variables that look like secrets
func (p *EnvProvider) List(_ context.Context) ([]Ref, error) {
	var refs []Ref

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		name := pair[0]
		value := pair[1]

		// Only include variables that look like secrets
		if isLikelySecretEnvVar(name, value) {
			refs = append(refs, Ref{
				Type:     "env",
				Name:     name,
				Original: fmt.Sprintf("${env:%s}", name),
			})
		}
	}

	return refs, nil
}

// IsAvailable always returns true as environment variables are always available
func (p *EnvProvider) IsAvailable() bool {
	return true
}

// isLikelySecretEnvVar returns true if the environment variable looks like it contains a secret
func isLikelySecretEnvVar(name, value string) bool {
	if value == "" {
		return false
	}

	// Check if the variable name suggests it's a secret
	isSecret, confidence := DetectPotentialSecret(value, name)

	// Lower threshold for env vars since they're commonly used for secrets
	return isSecret || confidence >= 0.3
}
