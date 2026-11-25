package source

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRedirectIsNotFollowed(t *testing.T) {
	t.Parallel()

	var targetRequests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetRequests.Add(1)
	}))
	defer target.Close()

	location := target.URL + "/redirect-target"
	var sourceRequests atomic.Int32
	source := startServer(t, func(w http.ResponseWriter, r *http.Request) {
		sourceRequests.Add(1)
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusFound)
	})
	h := clientFor(t, source, Options{})

	operations := []struct {
		name string
		call func() error
	}{
		{
			name: "Account",
			call: func() error {
				_, err := h.Account(context.Background(), testAccount)
				return err
			},
		},
		{
			name: "Network",
			call: func() error {
				_, err := h.Network(context.Background())
				return err
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			err := op.call()
			if !errors.Is(err, ErrStatus) {
				t.Fatalf("error = %v, want ErrStatus", err)
			}
			if strings.Contains(err.Error(), target.URL) || strings.Contains(err.Error(), location) {
				t.Fatalf("error exposes redirect target: %v", err)
			}
		})
	}

	if got := sourceRequests.Load(); got != int32(len(operations)) {
		t.Errorf("source listener requests = %d, want %d", got, len(operations))
	}
	if got := targetRequests.Load(); got != 0 {
		t.Errorf("redirect target requests = %d, want 0", got)
	}
}
