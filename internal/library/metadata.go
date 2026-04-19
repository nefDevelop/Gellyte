package library

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// FFProbeResult representa la salida simplificada de ffprobe
type FFProbeResult struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
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
}

// GetVideoMetadata ejecuta ffprobe para extraer información técnica de un archivo de video.
func GetVideoMetadata(path string) (*VideoMetadata, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration,bit_rate:stream=codec_name,width,height",
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

	// Extraer datos del primer stream de video
	for _, stream := range result.Streams {
		if stream.Width > 0 { // Es un stream de video
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.VideoCodec = stream.CodecName
			break
		}
	}

	return metadata, nil
}

// GenerateThumbnail extrae un fotograma del video para usarlo como miniatura.
func GenerateThumbnail(videoPath, thumbPath string) error {
	// Extraer un fotograma al primer segundo (más seguro para clips cortos)
	cmd := exec.Command("ffmpeg",
		"-ss", "00:00:01",
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		"-y", // Sobrescribir si existe
		thumbPath)

	return cmd.Run()
}
