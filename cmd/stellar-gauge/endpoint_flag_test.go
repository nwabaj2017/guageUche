package main

import (
	"strings"
	"testing"
)

func TestDuplicateEndpointDoesNotEchoTheValue(t *testing.T) {
	t.Parallel()

	t.Run("two duplicate endpoints refuse with exit 2 and do not echo userinfo or host", func(t *testing.T) {
		t.Parallel()
		code, stdout, stderr := execCLI(t,
			"check",
			"--network", testNetwork,
			"--endpoint", "https://user:pass@a.invalid",
			"--endpoint", "https://user:pass@a.invalid",
			accountOne,
		)
		if code != 2 {
			t.Fatalf("exit code = %d, want 2 (stderr %q)", code, stderr)
		}
		for _, secret := range []string{"user:pass", "a.invalid", "b.invalid"} {
			if strings.Contains(stdout, secret) {
				t.Errorf("stdout echoes secret %q: %s", secret, stdout)
			}
			if strings.Contains(stderr, secret) {
				t.Errorf("stderr echoes secret %q: %s", secret, stderr)
			}
		}
	})

	t.Run("three endpoints with duplicate refuse with exit 2 and do not echo", func(t *testing.T) {
		t.Parallel()
		code, stdout, stderr := execCLI(t,
			"check",
			"--network", testNetwork,
			"--endpoint", "https://user:pass@a.invalid",
			"--endpoint", "https://b.invalid",
			"--endpoint", "https://user:pass@a.invalid",
			accountOne,
		)
		if code != 2 {
			t.Fatalf("exit code = %d, want 2 (stderr %q)", code, stderr)
		}
		for _, secret := range []string{"user:pass", "a.invalid", "b.invalid"} {
			if strings.Contains(stdout, secret) {
				t.Errorf("stdout echoes secret %q: %s", secret, stdout)
			}
			if strings.Contains(stderr, secret) {
				t.Errorf("stderr echoes secret %q: %s", secret, stderr)
			}
		}
	})

	t.Run("endpoint error names public endpoints without echoing", func(t *testing.T) {
		t.Parallel()
		code, stdout, stderr := execCLI(t,
			"check",
			"--network", testNetwork,
			accountOne,
		)
		if code != 2 {
			t.Fatalf("exit code = %d, want 2 (stderr %q)", code, stderr)
		}
		if !strings.Contains(stderr, "https://horizon.stellar.org") {
			t.Errorf("stderr does not name public mainnet endpoint: %s", stderr)
		}
		if !strings.Contains(stderr, "https://horizon-testnet.stellar.org") {
			t.Errorf("stderr does not name public testnet endpoint: %s", stderr)
		}
		if stdout != "" {
			t.Errorf("stdout is not empty: %s", stdout)
		}
	})
}
