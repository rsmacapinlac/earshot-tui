//go:build linux

package platform

import (
	"fmt"
	"os/exec"
	"sync"
)

type linuxAudioPlayer struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

// NewAudioPlayer returns an AudioPlayer that uses ffplay on Linux.
func NewAudioPlayer() AudioPlayer {
	return &linuxAudioPlayer{}
}

func (p *linuxAudioPlayer) Play(filePath string) error {
	if _, err := exec.LookPath("ffplay"); err != nil {
		return fmt.Errorf("ffplay not found; install ffmpeg: sudo apt install ffmpeg")
	}
	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", filePath)
	p.mu.Lock()
	p.cmd = cmd
	p.mu.Unlock()
	return cmd.Run()
}

func (p *linuxAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}
