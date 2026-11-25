package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/posture-sec/guage/internal/model"
	"github.com/posture-sec/guage/internal/source"
)

type blockingSource struct {
	entered chan struct{}
}

func (s *blockingSource) Network(ctx context.Context) (model.NetworkInfo, error) {
	if s.entered != nil {
		select {
		case s.entered <- struct{}{}:
		default:
		}
	}
	<-ctx.Done()
	return model.NetworkInfo{}, ctx.Err()
}

func (s *blockingSource) Account(_ context.Context, id string) (model.AccountSnapshot, error) {
	return model.AccountSnapshot{ID: id}, nil
}

func TestInterruptCancelsTheRun(t *testing.T) {
	t.Run("cancelled check exits 3, writes nothing to stdout, prints one line to stderr", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		entered := make(chan struct{}, 1)
		src := &blockingSource{entered: entered}
		d := offlineDeps(nil)
		d.ctx = ctx
		d.newSource = func(string) (source.Source, error) { return src, nil }

		var stdout, stderr bytes.Buffer
		done := make(chan int, 1)
		args := []string{"check", "--network", testNetwork, "--endpoint", "https://horizon.example", accountOne}
		go func() {
			done <- runCLI(args, &stdout, &stderr, d)
		}()

		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for Network to be entered")
		}

		cancel()

		select {
		case code := <-done:
			if code != 3 {
				t.Fatalf("exit code = %d, want 3", code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout is not empty: %q", stdout.String())
			}
			errStr := strings.TrimRight(stderr.String(), "\r\n")
			if strings.Contains(errStr, "\n") || len(errStr) == 0 {
				t.Errorf("stderr does not contain exactly one line: %q", stderr.String())
			}
			if !strings.HasPrefix(errStr, "stellar-gauge:") {
				t.Errorf("stderr does not have stellar-gauge prefix: %q", errStr)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for check to exit after cancel")
		}
	})

	t.Run("completed run still writes its report", func(t *testing.T) {
		d := offlineDeps(nil)
		d.newSource = func(string) (source.Source, error) { return completedSource{}, nil }
		code, stdout, stderr := execCLIWith(t, d, "check", "--network", testNetwork, "--endpoint", "https://horizon.example", accountOne)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, stderr)
		}
		if len(stdout) == 0 {
			t.Errorf("stdout is empty on completed run")
		}
	})
}

// completedSource answers both calls a check makes, so a run can finish without a
// socket. The interrupt test above blocks in Network; this one proves the signal
// handler is installed without turning a run that never sees a signal into a
// dial to the endpoint named on the command line.
type completedSource struct{}

func (completedSource) Network(context.Context) (model.NetworkInfo, error) {
	return model.NetworkInfo{Passphrase: testPassphrase, Ledger: 1}, nil
}

func (completedSource) Account(_ context.Context, id string) (model.AccountSnapshot, error) {
	return model.AccountSnapshot{
		ID:       id,
		Sequence: "1",
		Ledger:   1,
		Thresholds: model.Thresholds{
			Low: 2, Medium: 2, High: 2,
		},
		Signers: []model.Signer{
			{Key: id, Type: model.SignerEd25519, Weight: 1},
			{Key: "GBQHXYDEJD6NEIUQ5NYDBVNTRCACJXMLO32JUQHBGO524NMVYOMIN5CP", Type: model.SignerEd25519, Weight: 1},
		},
	}, nil
}
