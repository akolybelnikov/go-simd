package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrivateWorkerRejectsInvalidRequest(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"_worker"}, strings.NewReader("{}"), &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "invalid worker request") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestPublicCommandsRemainClosed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, strings.NewReader(""), &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "not enabled") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}
