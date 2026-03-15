//go:build darwin

package main

import (
	"fmt"
	"os"
	"strings"
)

// wrapCoreLaunchWithShell wraps the core launch with user's shell on macOS
func wrapCoreLaunchWithShell(coreBinary string, args []string) (string, []string, error) {
	shellPath, err := selectUserShell()
	if err != nil {
		return "", nil, err
	}

	command := buildShellExecCommand(coreBinary, args)
	return shellPath, []string{"-l", "-c", command}, nil
}

// selectUserShell finds the user's preferred shell
func selectUserShell() (string, error) {
	candidates := []string{}
	if shellEnv := strings.TrimSpace(os.Getenv("SHELL")); shellEnv != "" {
		candidates = append(candidates, shellEnv)
	}
	candidates = append(candidates,
		"/bin/zsh",
		"/bin/bash",
		"/bin/sh",
	)

	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}

		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no usable shell found for core launch")
}

// buildShellExecCommand builds a shell command with exec
func buildShellExecCommand(binary string, args []string) string {
	quoted := make([]string, 0, len(args)+1)
	quoted = append(quoted, shellQuote(binary))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}

	return "exec " + strings.Join(quoted, " ")
}

// shellQuote quotes an argument for POSIX shell
func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}

	var builder strings.Builder
	builder.Grow(len(arg) + 2)

	hasSpecial := false
	for _, r := range arg {
		switch r {
		case ' ', '\t', '\n', '\\', '"', '\'', ';', '&', '|', '(', ')', '<', '>', '$', '`', '!', '*', '?', '[', ']', '{', '}', '~', '#', '%', '=':
			hasSpecial = true
		}
	}

	if !hasSpecial {
		return arg
	}

	builder.WriteByte('\'')
	for _, r := range arg {
		if r == '\'' {
			builder.WriteString(`'"'"'`)
		} else {
			builder.WriteRune(r)
		}
	}
	builder.WriteByte('\'')

	return builder.String()
}
