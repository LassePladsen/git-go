package object

import (
	"io/fs"
	"mygit/test"
	"testing"
)

func TestFileModeToGitMode(t *testing.T) {
	if mode, err := fileModeToGitMode(fs.ModeNamedPipe); err == nil {
		t.Errorf("Expected error for unsupported mode %T, got nil. Mode output was: %v", fs.ModeNamedPipe, mode)
	}
	
	want := "120000" 
	got, err := fileModeToGitMode(fs.ModeSymlink)
	if err != nil {
		t.Errorf("Expected: %v, got error: %v", want, err)
	}
	test.Expect(t, got, want)
}
