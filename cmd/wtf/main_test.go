package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"strings"
	"testing"
)

// TestRun_Help verifies that a help request prints usage to the injected writer
// and reports flag.ErrHelp (a successful, non-error invocation).
func TestRun_Help(t *testing.T) {
	var buf bytes.Buffer
	err := Run(context.Background(), []string{"help"}, &buf)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("err=%v, want flag.ErrHelp", err)
	}
	if got := buf.String(); !strings.Contains(got, "wtf <command>") {
		t.Fatalf("usage not written to stdout; got %q", got)
	}
}

// TestRun_UnknownCommand verifies an unknown command returns an error and writes
// nothing to stdout.
func TestRun_UnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	err := Run(context.Background(), []string{"bogus"}, &buf)
	if err == nil || !strings.Contains(err.Error(), "wtf bogus: unknown command") {
		t.Fatalf("err=%v, want unknown command error", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", buf.String())
	}
}

// TestDialCommand_Help verifies the dial subcommand dispatcher writes its usage
// to the injected writer.
func TestDialCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	err := (&DialCommand{}).Run(context.Background(), []string{"help"}, &buf)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("err=%v, want flag.ErrHelp", err)
	}
	if got := buf.String(); !strings.Contains(got, "wtf dial <command>") {
		t.Fatalf("dial usage not written to stdout; got %q", got)
	}
}
