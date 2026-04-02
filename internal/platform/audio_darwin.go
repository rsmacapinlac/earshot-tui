//go:build darwin

package platform

import (
	"os/exec"
	"sync"
)

type darwinAudioPlayer struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

// NewAudioPlayer returns an AudioPlayer that uses afplay on macOS.
func NewAudioPlayer() AudioPlayer {
	return &darwinAudioPlayer{}
}

func (p *darwinAudioPlayer) Play(filePath string) error {
	cmd := exec.Command("afplay", filePath)
	p.mu.Lock()
	p.cmd = cmd
	p.mu.Unlock()
	return cmd.Run()
}

func (p *darwinAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}
