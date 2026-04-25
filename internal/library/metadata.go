package library

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
)

// FFProbeResult representa la salida simplificada de ffprobe
type FFProbeResult struct {
	Streams []struct {
		Index       int               `json:"index"`
		CodecType   string            `json:"codec_type"`
		CodecName   string            `json:"codec_name"`
		Width       int               `json:"width"`
		Height      int               `json:"height"`
		Tags        map[string]string `json:"tags"`
		Disposition map[string]int    `json:"disposition"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

// VideoMetadata contiene la información técnica extraída
type VideoMetadata struct {
	DurationTicks int64
	Width         int
	Height        int
	Bitrate       int64
	VideoCodec    string
	AudioCodec    string
	Streams       []models.MediaStream
}

// GetRawMetadata ejecuta ffprobe y devuelve la estructura completa.
func GetRawMetadata(path string) (*FFProbeResult, error) {
	cmd := exec.Command(config.AppConfig.Transcoder.FFprobePath,
		"-v", "error",
		"-show_entries", "format=duration,bit_rate:stream=index,codec_type,codec_name,width,height:stream_tags=language,title:stream_disposition=default",
		"-of", "json=c=1",
		path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %v (output: %s)", err, string(output))
	}

	var result FFProbeResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("error parseando JSON de ffprobe: %v", err)
	}
	return &result, nil
}

// GetVideoMetadata ejecuta ffprobe para extraer información técnica de un archivo de video.
func GetVideoMetadata(path string) (*VideoMetadata, error) {
	resultPtr, err := GetRawMetadata(path)
	if err != nil {
		return nil, err
	}
	result := *resultPtr

	metadata := &VideoMetadata{}

	// Extraer duración (ffprobe devuelve segundos en string)
	if result.Format.Duration != "" {
		durationSec, _ := strconv.ParseFloat(result.Format.Duration, 64)
		// Jellyfin usa ticks (1 segundo = 10,000,000 ticks)
		metadata.DurationTicks = int64(durationSec * 10000000)
	}

	// Extraer bitrate
	if result.Format.BitRate != "" {
		metadata.Bitrate, _ = strconv.ParseInt(result.Format.BitRate, 10, 64)
	}

	// Extraer datos de los flujos (video, audio, subtítulos)
	for _, s := range result.Streams {
		streamType := ""
		switch s.CodecType {
		case "video":
			streamType = "Video"
			if metadata.VideoCodec == "" {
				metadata.Width = s.Width
				metadata.Height = s.Height
				metadata.VideoCodec = s.CodecName
			}
		case "audio":
			streamType = "Audio"
			if metadata.AudioCodec == "" {
				metadata.AudioCodec = s.CodecName
			}
		case "subtitle":
			streamType = "Subtitle"
		}

		if streamType != "" {
			lang := ""
			if s.Tags != nil {
				lang = s.Tags["language"]
			}

			isDefault := false
			if s.Disposition != nil && s.Disposition["default"] == 1 {
				isDefault = true
			}

			metadata.Streams = append(metadata.Streams, models.MediaStream{
				Type:      streamType,
				Index:     s.Index,
				Codec:     s.CodecName,
				Language:  lang,
				IsDefault: isDefault,
				Width:     s.Width,
				Height:    s.Height,
			})
		}
	}

	return metadata, nil
}

// GenerateThumbnail extrae un fotograma del video para usarlo como miniatura.
func GenerateThumbnail(videoPath, thumbPath string) error {
	// Extraer un fotograma al primer segundo (más seguro para clips cortos)
	cmd := exec.Command(config.AppConfig.Transcoder.FFmpegPath,
		"-ss", "00:00:01",
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-y", // Sobrescribir si existe
		thumbPath)

	return cmd.Run()
}
