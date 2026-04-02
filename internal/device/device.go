// Package device handles scanning and downloading from earshot device filesystems.
package device

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/rsmacapinlac/earshot-tui/internal/recording"
)

// recordingDirPattern matches earshot folder names: 20260401T083114
var recordingDirPattern = regexp.MustCompile(`^\d{8}T\d{6}$`)

// folderTimeLayout is the Go time layout for earshot folder names.
const folderTimeLayout = "20060102T150405"

// Folder represents a recording folder found on an earshot device.
type Folder struct {
	Name      string    // e.g. "20260401T083114"
	Path      string    // full path on device filesystem
	Count     int       // number of .opus files
	Duration  float64   // total duration in seconds (0 if ffprobe unavailable)
	Timestamp time.Time // parsed from folder name; zero if unparseable
}

// Scan reads the recording folders directly from the device mount root and
// returns all folders that contain at least one .opus file, sorted most-recent-first.
func Scan(mountPath string) ([]Folder, error) {
	entries, err := os.ReadDir(mountPath)
	if err != nil {
		return nil, fmt.Errorf("read device at %s: %w", mountPath, err)
	}

	var folders []Folder
	for _, e := range entries {
		if !e.IsDir() || !recordingDirPattern.MatchString(e.Name()) {
			continue
		}
		dir := filepath.Join(mountPath, e.Name())
		opus := findOpusFiles(dir)
		if len(opus) == 0 {
			continue
		}
		t, _ := time.ParseInLocation(folderTimeLayout, e.Name(), time.Local)
		folders = append(folders, Folder{
			Name:      e.Name(),
			Path:      dir,
			Count:     len(opus),
			Duration:  totalDuration(opus),
			Timestamp: t,
		})
	}

	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Timestamp.Before(folders[j].Timestamp)
	})
	return folders, nil
}

// Download copies audio from src (a device Folder) into a new local recording
// directory under recordingsDir. Progress is reported as a fraction 0.0–1.0 via
// progressFn. Returns the local recording directory path.
//
// When a folder contains multiple .opus files they are concatenated in filename
// order into a single recording.opus using ffmpeg's concat demuxer.
func Download(src Folder, deviceName string, recordingsDir string, progressFn func(float64)) (string, error) {
	opusFiles := findOpusFiles(src.Path)
	if len(opusFiles) == 0 {
		return "", fmt.Errorf("no .opus files in %s", src.Path)
	}

	localDir := filepath.Join(recordingsDir, src.Name)
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return "", fmt.Errorf("create local dir %s: %w", localDir, err)
	}

	dest := recording.AudioPath(localDir)
	var copyErr error
	if len(opusFiles) == 1 {
		copyErr = copyWithProgress(opusFiles[0], dest, progressFn)
	} else {
		copyErr = concatWithProgress(opusFiles, dest, progressFn)
	}
	if copyErr != nil {
		os.RemoveAll(localDir)
		return "", copyErr
	}

	now := time.Now().UTC()
	status := &recording.Status{
		Status:       recording.StateDownloaded,
		Device:       deviceName,
		RecordedAt:   src.Timestamp,
		Duration:     src.Duration,
		DownloadedAt: &now,
	}
	if err := recording.SaveStatus(localDir, status); err != nil {
		os.RemoveAll(localDir)
		return "", fmt.Errorf("save status: %w", err)
	}

	return localDir, nil
}

// IsImported reports whether deviceFolderName has already been downloaded
// into recordingsDir.
func IsImported(deviceFolderName, recordingsDir string) bool {
	localDir := filepath.Join(recordingsDir, deviceFolderName)
	_, err := os.Stat(filepath.Join(localDir, "status.json"))
	return err == nil
}

// Delete removes a recording folder from the device.
func Delete(src Folder) error {
	return os.RemoveAll(src.Path)
}

// ---- helpers ----------------------------------------------------------------

func findOpusFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".opus" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // audio-001 before audio-002
	return files
}

func totalDuration(files []string) float64 {
	var total float64
	for _, f := range files {
		if d, err := probeDuration(f); err == nil {
			total += d
		}
	}
	return total
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

type ffprobeOutput struct {
	Format ffprobeFormat `json:"format"`
}

func probeDuration(path string) (float64, error) {
	out, err := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		path,
	).Output()
	if err != nil {
		return 0, err
	}
	var fo ffprobeOutput
	if err := json.Unmarshal(out, &fo); err != nil {
		return 0, err
	}
	var d float64
	fmt.Sscanf(fo.Format.Duration, "%f", &d)
	return d, nil
}

// concatWithProgress concatenates multiple .opus files into dst using ffmpeg's
// concat demuxer, reporting coarse progress (one step per file copied).
func concatWithProgress(files []string, dst string, progressFn func(float64)) error {
	// Write a temporary concat list file.
	listFile, err := os.CreateTemp("", "earshot-concat-*.txt")
	if err != nil {
		return fmt.Errorf("create concat list: %w", err)
	}
	defer os.Remove(listFile.Name())

	for _, f := range files {
		fmt.Fprintf(listFile, "file '%s'\n", f)
	}
	listFile.Close()

	progressFn(0.1) // signal that work has started
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile.Name(),
		"-c", "copy",
		dst,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w\n%s", err, out)
	}
	progressFn(1.0)
	return nil
}

// copyWithProgress copies src to dst while reporting byte-level progress.
func copyWithProgress(src, dst string, progressFn func(float64)) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	total := info.Size()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	pr := &progressReader{r: in, total: total, fn: progressFn}
	if _, err := io.Copy(out, pr); err != nil {
		return err
	}
	return out.Sync()
}

type progressReader struct {
	r       io.Reader
	total   int64
	written int64
	fn      func(float64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.written += int64(n)
	if pr.total > 0 {
		pr.fn(float64(pr.written) / float64(pr.total))
	}
	return n, err
}
