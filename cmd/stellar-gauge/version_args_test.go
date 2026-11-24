package main

import (
	"strings"
	"testing"
)

func TestVersionRejectsArguments(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "bare version",
			args:       []string{"version"},
			wantCode:   0,
			wantStdout: "stellar-gauge " + version + "\n",
		},
		{
			name:       "version --json",
			args:       []string{"version", "--json"},
			wantCode:   2,
			wantStderr: "stellar-gauge: version takes no arguments\nusage: stellar-gauge",
		},
		{
			name:       "version garbage",
			args:       []string{"version", "garbage"},
			wantCode:   2,
			wantStderr: "stellar-gauge: version takes no arguments\nusage: stellar-gauge",
		},
		{
			name:       "version version",
			args:       []string{"version", "version"},
			wantCode:   2,
			wantStderr: "stellar-gauge: version takes no arguments\nusage: stellar-gauge",
		},
		{
			name:       "help unaffected",
			args:       []string{"help"},
			wantCode:   0,
			wantStdout: "usage: stellar-gauge",
		},
		{
			name:       "-h unaffected",
			args:       []string{"-h"},
			wantCode:   0,
			wantStdout: "usage: stellar-gauge",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, stdout, stderr := execCLI(t, tc.args...)
			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr %q)", code, tc.wantCode, stderr)
			}
			if tc.wantStdout != "" && !strings.Contains(stdout, tc.wantStdout) {
				t.Errorf("stdout %q does not contain %q", stdout, tc.wantStdout)
			}
			if tc.wantStderr != "" && !strings.Contains(stderr, tc.wantStderr) {
				t.Errorf("stderr %q does not contain %q", stderr, tc.wantStderr)
			}
		})
	}
}
