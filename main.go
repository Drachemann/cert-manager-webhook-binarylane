package main

import (
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/drachemann/cert-manager-webhook-binarylane/pkg/solver"
)

func main() {
	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		groupName = "acme.binarylane.com"
	}
	cmd.RunWebhookServer(groupName, &solver.Solver{})
}
