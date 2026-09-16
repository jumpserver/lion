package tunnel

import (
	"os"
	"testing"

	"github.com/jumpserver-dev/sdk-go/model"
)

func TestVideoWorkerRequiresValidEnterpriseLicense(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		terminal *model.TerminalConfig
		want     bool
	}{
		{name: "switch off", terminal: &model.TerminalConfig{LicenseIsValid: true}},
		{name: "community edition", enabled: true, terminal: &model.TerminalConfig{}},
		{name: "missing terminal config", enabled: true},
		{name: "licensed and enabled", enabled: true, terminal: &model.TerminalConfig{LicenseIsValid: true}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldUseVideoWorker(test.enabled, test.terminal); got != test.want {
				t.Fatalf("shouldUseVideoWorker() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestMissingTerminalConfigRetainsReplay(t *testing.T) {
	root := t.TempDir()
	uploader := PartUploader{SessionId: "session-id", RootPath: root}
	uploader.uploadToStorage(root)
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("source directory was not retained for recovery: %v", err)
	}
}
