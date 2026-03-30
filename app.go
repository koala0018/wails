package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

type FileSelection struct {
	Path string `json:"path"`
}

type SplitRequest struct {
	InputPath        string `json:"inputPath"`
	OutputDir        string `json:"outputDir"`
	SegmentLengthSec int    `json:"segmentLengthSec"`
}

type SplitResult struct {
	OutputDir      string   `json:"outputDir"`
	SegmentCount   int      `json:"segmentCount"`
	GeneratedFiles []string `json:"generatedFiles"`
	Command        string   `json:"command"`
}

type EnvironmentStatus struct {
	FFmpegReady bool   `json:"ffmpegReady"`
	FFmpegPath  string `json:"ffmpegPath"`
	Message     string `json:"message"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectVideoFile() (*FileSelection, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a video file",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.mov;*.mkv;*.avi;*.flv;*.wmv;*.m4v;*.webm",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if selection == "" {
		return &FileSelection{}, nil
	}
	return &FileSelection{Path: selection}, nil
}

func (a *App) SelectOutputDirectory() (*FileSelection, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select output folder",
	})
	if err != nil {
		return nil, err
	}
	if selection == "" {
		return &FileSelection{}, nil
	}
	return &FileSelection{Path: selection}, nil
}

func (a *App) GetEnvironmentStatus() EnvironmentStatus {
	ffmpegPath, err := findFFmpeg()
	if err != nil {
		return EnvironmentStatus{
			FFmpegReady: false,
			Message:     "ffmpeg not found. Video splitting is unavailable.",
		}
	}

	return EnvironmentStatus{
		FFmpegReady: true,
		FFmpegPath:  ffmpegPath,
		Message:     "ffmpeg is ready.",
	}
}

func (a *App) SplitVideo(request SplitRequest) (*SplitResult, error) {
	if strings.TrimSpace(request.InputPath) == "" {
		return nil, errors.New("please select a source video")
	}
	if strings.TrimSpace(request.OutputDir) == "" {
		return nil, errors.New("please select an output folder")
	}
	if request.SegmentLengthSec <= 0 {
		return nil, errors.New("segment length must be greater than 0 seconds")
	}

	inputPath, err := filepath.Abs(request.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve input path: %w", err)
	}
	outputDir, err := filepath.Abs(request.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve output path: %w", err)
	}

	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("input video is unavailable: %w", err)
	}
	if inputInfo.IsDir() {
		return nil, errors.New("input path must be a file")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output folder: %w", err)
	}

	ffmpegPath, err := findFFmpeg()
	if err != nil {
		return nil, errors.New("ffmpeg was not found")
	}
	ffprobePath, err := findFFprobe(ffmpegPath)
	if err != nil {
		return nil, errors.New("ffprobe was not found")
	}

	durationSeconds, err := probeDurationSeconds(a.ctx, ffprobePath, inputPath)
	if err != nil {
		return nil, err
	}

	baseName := sanitizeBaseName(strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)))
	pattern := filepath.Join(outputDir, fmt.Sprintf("%s_part_*.mp4", baseName))

	existingFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan existing output files: %w", err)
	}
	for _, existingFile := range existingFiles {
		if removeErr := os.Remove(existingFile); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to remove old output file %s: %w", existingFile, removeErr)
		}
	}

	targetCount := int(math.Ceil(durationSeconds / float64(request.SegmentLengthSec)))
	if targetCount <= 0 {
		return nil, errors.New("video duration is invalid")
	}

	commandPreview := ""
	for i := 0; i < targetCount; i++ {
		startSeconds := float64(i * request.SegmentLengthSec)
		currentDuration := math.Min(float64(request.SegmentLengthSec), durationSeconds-startSeconds)
		if currentDuration <= 0 {
			break
		}

		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_part_%03d.mp4", baseName, i+1))
		args := []string{
			"-hide_banner",
			"-y",
			"-ss", formatFFmpegSeconds(startSeconds),
			"-i", inputPath,
			"-t", formatFFmpegSeconds(currentDuration),
			"-map", "0:v:0",
			"-map", "0:a?",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "128k",
			"-movflags", "+faststart",
			outputPath,
		}

		if commandPreview == "" {
			commandPreview = ffmpegPath + " " + strings.Join(args, " ")
		}

		cmd := exec.CommandContext(a.ctx, ffmpegPath, args...)
		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			message := strings.TrimSpace(string(output))
			if message == "" {
				message = runErr.Error()
			}
			return nil, fmt.Errorf("segment %d failed: %s", i+1, message)
		}
	}

	generatedFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("split completed, but failed to read output files: %w", err)
	}
	if len(generatedFiles) == 0 {
		return nil, errors.New("split completed, but no output files were created")
	}
	sort.Strings(generatedFiles)

	return &SplitResult{
		OutputDir:      outputDir,
		SegmentCount:   len(generatedFiles),
		GeneratedFiles: generatedFiles,
		Command:        commandPreview,
	}, nil
}

func findFFmpeg() (string, error) {
	if configuredPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); configuredPath != "" {
		if _, err := os.Stat(configuredPath); err == nil || errors.Is(err, os.ErrPermission) {
			return configuredPath, nil
		}
	}

	candidates := []string{
		"ffmpeg",
		filepath.Join("tools", "ffmpeg", "bin", "ffmpeg.exe"),
		filepath.Join("ffmpeg", "bin", "ffmpeg.exe"),
		`C:\ffmpeg\bin\ffmpeg.exe`,
		`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates,
			filepath.Join(localAppData, "Microsoft", "WinGet", "Links", "ffmpeg.exe"),
			filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "Gyan.FFmpeg_Microsoft.Winget.Source_8wekyb3d8bbwe", "ffmpeg-8.1-full_build", "bin", "ffmpeg.exe"),
		)
	}

	for _, candidate := range candidates {
		if candidate == "ffmpeg" {
			if path, err := exec.LookPath(candidate); err == nil {
				return path, nil
			}
			continue
		}
		if _, err := os.Stat(candidate); err == nil || errors.Is(err, os.ErrPermission) {
			return candidate, nil
		}
	}

	return "", errors.New("ffmpeg not found")
}

func findFFprobe(ffmpegPath string) (string, error) {
	if configuredPath := strings.TrimSpace(os.Getenv("FFPROBE_PATH")); configuredPath != "" {
		if _, err := os.Stat(configuredPath); err == nil || errors.Is(err, os.ErrPermission) {
			return configuredPath, nil
		}
	}

	ffprobeCandidate := filepath.Join(filepath.Dir(ffmpegPath), "ffprobe.exe")
	if _, err := os.Stat(ffprobeCandidate); err == nil || errors.Is(err, os.ErrPermission) {
		return ffprobeCandidate, nil
	}

	if path, err := exec.LookPath("ffprobe"); err == nil {
		return path, nil
	}

	return "", errors.New("ffprobe not found")
}

func probeDurationSeconds(ctx context.Context, ffprobePath, inputPath string) (float64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return 0, fmt.Errorf("failed to read video duration: %s", message)
	}

	durationText := strings.TrimSpace(string(output))
	durationSeconds, err := strconv.ParseFloat(durationText, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse video duration: %w", err)
	}
	if durationSeconds <= 0 {
		return 0, errors.New("video duration must be greater than 0")
	}

	return durationSeconds, nil
}

func formatFFmpegSeconds(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}

func sanitizeBaseName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "video"
	}

	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	safeName := invalidChars.ReplaceAllString(trimmed, "_")
	safeName = strings.Trim(safeName, ". ")
	if safeName == "" {
		return "video"
	}

	return safeName
}
