package transcoder

import (
	"strings"
	"testing"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
)

func TestBuildTranscodeCmd(t *testing.T) {
	config.AppConfig.Transcoder.FFmpegPath = "ffmpeg"
	item := models.MediaItem{
		Path:   "/path/to/video.mp4",
		Height: 1080,
	}
	opts := TranscodeOptions{
		VideoCodec: "libx264",
		MaxHeight:  720,
		Bitrate:    2000000,
	}

	cmd := BuildTranscodeCmd(item, opts)

	if !strings.Contains(cmd.Path, "ffmpeg") {
		t.Errorf("expected ffmpeg in path, got %s", cmd.Path)
	}

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "-i /path/to/video.mp4") {
		t.Errorf("args missing input path: %s", args)
	}
	if !strings.Contains(args, "scale=-2:720") {
		t.Errorf("args missing scaling: %s", args)
	}
	if !strings.Contains(args, "-b:v 2000000") {
		t.Errorf("args missing bitrate: %s", args)
	}
}

func TestBuildHLSSegmentCmd(t *testing.T) {
	config.AppConfig.Transcoder.FFmpegPath = "ffmpeg"
	item := models.MediaItem{Path: "/path/to/video.mp4"}
	opts := TranscodeOptions{VideoCodec: "copy", AudioCodec: "copy"}

	cmd := BuildHLSSegmentCmd(item, 2, 10, opts)

	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "-ss 20") {
		t.Errorf("expected start time 20, got %s", args)
	}
	if !strings.Contains(args, "-t 10") {
		t.Errorf("expected duration 10, got %s", args)
	}
	if !strings.Contains(args, "-c:v copy") {
		t.Errorf("expected vcopy, got %s", args)
	}
}
