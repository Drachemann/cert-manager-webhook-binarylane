package main

import (
	"os"
	"testing"

	"github.com/drachemann/cert-manager-webhook-binarylane/pkg/solver"
)

func TestSolverImported(t *testing.T) {
	s := &solver.Solver{}
	if s.Name() != "binarylane" {
		t.Errorf("expected solver name 'binarylane', got %q", s.Name())
	}
}

func TestGroupNameEnv(t *testing.T) {
	os.Setenv("GROUP_NAME", "acme.binarylane.com")
	gn := os.Getenv("GROUP_NAME")
	if gn != "acme.binarylane.com" {
		t.Errorf("expected GROUP_NAME 'acme.binarylane.com', got %q", gn)
	}
}

