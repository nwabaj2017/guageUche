package source

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryAfterTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "empty"},
		{name: "HTTP date", value: "Wed, 21 Oct 2015 07:28:00 GMT"},
		{name: "zero", value: "0"},
		{name: "negative", value: "-3"},
		{name: "integer seconds", value: "3", want: 3 * time.Second},
		{name: "capped integer seconds", value: "600", want: 5 * time.Second},
		{name: "trimmed integer seconds", value: " 3 ", want: 3 * time.Second},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			header := make(http.Header)
			if tt.value != "" {
				header.Set("Retry-After", tt.value)
			}

			got := retryAfter(header)
			if got != tt.want {
				t.Fatalf("retryAfter(%q) = %v, want %v", tt.value, got, tt.want)
			}
			if got < 0 {
				t.Fatalf("retryAfter(%q) = %v, want a non-negative duration", tt.value, got)
			}
			if got > maxRetryAfter {
				t.Fatalf("retryAfter(%q) = %v, exceeds %v", tt.value, got, maxRetryAfter)
			}
		})
	}
}
