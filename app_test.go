package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSplitVideo(t *testing.T) {
	ffmpegPath, err := findFFmpeg()
	if err != nil {
		t.Fatalf("ffmpeg not available: %v", err)
	}

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.mp4")
	outputDir := filepath.Join(tempDir, "output")

	generateSample := exec.Command(
		ffmpegPath,
		"-hide_banner",
		"-y",
		"-f", "lavfi",
		"-i", "testsrc=size=640x360:rate=30",
		"-f", "lavfi",
		"-i", "sine=frequency=1000:sample_rate=44100",
		"-t", "7",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		inputPath,
	)
	if output, err := generateSample.CombinedOutput(); err != nil {
		t.Fatalf("failed to create sample video: %v\n%s", err, string(output))
	}

	app := NewApp()
	app.startup(context.Background())

	result, err := app.SplitVideo(SplitRequest{
		InputPath:        inputPath,
		OutputDir:        outputDir,
		SegmentLengthSec: 3,
	})
	if err != nil {
		t.Fatalf("SplitVideo returned error: %v", err)
	}

	if result.SegmentCount < 3 {
		t.Fatalf("expected at least 3 segments, got %d", result.SegmentCount)
	}

	if len(result.GeneratedFiles) != result.SegmentCount {
		t.Fatalf("expected generated file count to match segment count, got %d files and count %d", len(result.GeneratedFiles), result.SegmentCount)
	}
}
