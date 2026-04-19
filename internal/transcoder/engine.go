package transcoder

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/gellyte/gellyte/internal/models"
)

// TranscodeOptions define cómo se debe procesar el video.
type TranscodeOptions struct {
	VideoCodec     string
	AudioCodec     string
	Bitrate        int64
	MaxHeight      int
	StartTimeTicks int64
}

// BuildTranscodeCmd genera un comando FFmpeg para transcodificar en tiempo real hacia la salida estándar (stdout).
func BuildTranscodeCmd(item models.MediaItem, opts TranscodeOptions) *exec.Cmd {
	args := []string{
		"-v", "error",
	}

	// Posición de inicio (Seeking)
	if opts.StartTimeTicks > 0 {
		startTimeSec := float64(opts.StartTimeTicks) / 10000000.0
		args = append(args, "-ss", fmt.Sprintf("%.3f", startTimeSec))
	}

	args = append(args, "-i", item.Path)

	// Configuración de Video
	if opts.VideoCodec != "" {
		args = append(args, "-c:v", opts.VideoCodec)
	} else {
		args = append(args, "-c:v", "libx264") // Default seguro
	}

	// Ajuste de Resolución si es necesario
	if opts.MaxHeight > 0 && item.Height > opts.MaxHeight {
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", opts.MaxHeight))
	}

	// Bitrate
	if opts.Bitrate > 0 {
		args = append(args, "-b:v", strconv.FormatInt(opts.Bitrate, 10))
		args = append(args, "-maxrate", strconv.FormatInt(opts.Bitrate, 10))
		args = append(args, "-bufsize", strconv.FormatInt(opts.Bitrate*2, 10))
	}

	// Configuración de Audio
	if opts.AudioCodec != "" {
		args = append(args, "-c:a", opts.AudioCodec)
	} else {
		args = append(args, "-c:a", "aac") // Default universal
	}

	// Formato de salida (Matroska o MPEGTS para streaming continuo)
	args = append(args, "-f", "matroska", "-")

	return exec.Command("ffmpeg", args...)
}

// BuildHLSSegmentCmd genera un comando para extraer un segmento específico de 10 segundos.
func BuildHLSSegmentCmd(item models.MediaItem, segmentIndex int, segmentDuration int, opts TranscodeOptions) *exec.Cmd {
	startTime := segmentIndex * segmentDuration
	
	args := []string{
		"-v", "error",
		"-ss", strconv.Itoa(startTime),
		"-t", strconv.Itoa(segmentDuration),
		"-i", item.Path,
		"-copyts", // Mantener timestamps para sincronización
	}

	// Configuración de Video (usamos libx264 para máxima compatibilidad HLS)
	args = append(args, "-c:v", "libx264", "-preset", "veryfast")
	
	if opts.MaxHeight > 0 && item.Height > opts.MaxHeight {
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", opts.MaxHeight))
	}

	// Audio a AAC (estándar HLS)
	args = append(args, "-c:a", "aac", "-ac", "2", "-b:a", "128k")

	// Formato MPEG-TS para segmentos HLS
	args = append(args, "-f", "mpegts", "-")

	return exec.Command("ffmpeg", args...)
}
