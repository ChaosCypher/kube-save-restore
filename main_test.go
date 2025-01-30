package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/config"
	"github.com/chaoscypher/kube-save-restore/internal/logger"
	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	"github.com/chaoscypher/kube-save-restore/internal/compare"
)

// TestGetKubeconfigPath remains unchanged

func TestHandleCompare(t *testing.T) {
	mockLogger := &logger.MockLogger{}
	mockK8sClient := &kubernetes.Client{Clientset: fake.NewSimpleClientset()}

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "Valid compare configuration",
			config: &config.Config{
				CompareSource: "/path/to/source",
				CompareTarget: "/path/to/target",
				CompareType:   "all",
				DryRun:        true,
			},
			wantErr: false,
		},
		{
			name: "Missing compare source",
			config: &config.Config{
				CompareTarget: "/path/to/target",
				CompareType:   "all",
				DryRun:        true,
			},
			wantErr: true,
		},
		{
			name: "Missing compare target",
			config: &config.Config{
				CompareSource: "/path/to/source",
				CompareType:   "all",
				DryRun:        true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleCompare(tt.config, mockK8sClient, mockLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCompare() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if logger received expected messages
			if len(mockLogger.InfoMessages) < 1 {
				t.Errorf("Expected at least 1 info message, got %d", len(mockLogger.InfoMessages))
			}

			// Reset mock logger for next test
			mockLogger.InfoMessages = []string{}
			mockLogger.ErrorMessages = []string{}
		})
	}
}

func TestRun(t *testing.T) {
	mockLogger := &logger.MockLogger{}

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "Backup mode",
			config: &config.Config{
				Mode:      "backup",
				BackupDir: "/tmp/backup",
			},
			wantErr: false,
		},
		{
			name: "Restore mode",
			config: &config.Config{
				Mode:       "restore",
				RestoreDir: "/tmp/restore",
			},
			wantErr: false,
		},
		{
			name: "Compare mode",
			config: &config.Config{
				Mode:          "compare",
				CompareSource: "/path/to/source",
				CompareTarget: "/path/to/target",
				CompareType:   "all",
			},
			wantErr: false,
		},
		{
			name: "Invalid mode",
			config: &config.Config{
				Mode: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.config, mockLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset mock logger for next test
			mockLogger.InfoMessages = []string{}
			mockLogger.ErrorMessages = []string{}
		})
	}
}
