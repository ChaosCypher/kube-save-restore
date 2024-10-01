package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/config"
	"github.com/chaoscypher/kube-save-restore/internal/logger"
)

func TestGetKubeconfigPath(t *testing.T) {
	logger := logger.SetupLogger(&config.Config{})

	tests := []struct {
		name       string
		kubeconfig string
		want       string
		setup      func()
		teardown   func()
	}{
		{
			name:       "With provided kubeconfig",
			kubeconfig: "/path/to/kubeconfig",
			want:       "/path/to/kubeconfig",
		},
		{
			name:       "Without provided kubeconfig",
			kubeconfig: "",
			want:       filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
			}

			got := getKubeconfigPath(tt.kubeconfig, logger)
			if got != tt.want {
				t.Errorf("getKubeconfigPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
