package transcoder

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
)

// TranscodeOptions define cómo se debe procesar el video.
type TranscodeOptions struct {
	VideoCodec          string
	AudioCodec          string
	Bitrate             int64
	MaxHeight           int
	StartTimeTicks      int64
	AudioStreamIndex    int
	SubtitleStreamIndex int
}

// BuildTranscodeCmd genera un comando FFmpeg para transcodificar en tiempo real hacia la salida estándar (stdout).
func BuildTranscodeCmd(item models.MediaItem, opts TranscodeOptions) *exec.Cmd {
	args := []string{
		"-v", "error",
		"-threads", "2", // Limita el número de hilos de CPU para reducir drásticamente el consumo de RAM
	}

	// Optimizaciones de hardware
	args = append(args, getHWAccelArgs()...)

	// Posición de inicio (Seeking)
	if opts.StartTimeTicks > 0 {
		startTimeSec := float64(opts.StartTimeTicks) / 10000000.0
		args = append(args, "-ss", fmt.Sprintf("%.3f", startTimeSec))
	}

	args = append(args, "-i", item.Path)

	// Selección de Pistas (Mapping)
	args = append(args, "-map", "0:v:0") // Tomar siempre el primer video de origen

	if opts.AudioStreamIndex > 0 { // Jellyfin envía el índice absoluto del stream (ej. 1, 2, 3)
		args = append(args, "-map", fmt.Sprintf("0:%d", opts.AudioStreamIndex))
	} else {
		args = append(args, "-map", "0:a:0?") // Fallback al primer audio disponible
	}

	// Configuración de Video
	if opts.VideoCodec != "" {
		args = append(args, "-c:v", opts.VideoCodec)
	} else {
		args = append(args, "-c:v", "libx264") // Default seguro
	}

	// FFmpeg da error si intentas escalar o cambiar bitrate mientras haces "copy" (Remux)
	if opts.VideoCodec != "copy" {
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

		// Optimización estricta para streaming en tiempo real sin cortes
		if opts.VideoCodec == "libx264" || opts.VideoCodec == "" {
			args = append(args, "-preset", "veryfast", "-tune", "zerolatency")
		}
	}

	// Configuración de Audio
	if opts.AudioCodec != "" {
		args = append(args, "-c:a", opts.AudioCodec)
	} else {
		args = append(args, "-c:a", "aac") // Default universal
	}

	// Si recodificamos audio, asegurar que sea estéreo (seguro para navegadores web)
	if opts.AudioCodec != "copy" {
		args = append(args, "-ac", "2")
	}

	// Formato de salida (Matroska o MPEGTS para streaming continuo)
	args = append(args, "-f", "matroska", "-")

	return exec.Command(config.AppConfig.Transcoder.FFmpegPath, args...)
}

// BuildHLSSegmentCmd genera un comando para extraer un segmento específico de 10 segundos.
func BuildHLSSegmentCmd(item models.MediaItem, segmentIndex int, segmentDuration int, opts TranscodeOptions) *exec.Cmd {
	startTime := segmentIndex * segmentDuration

	args := []string{
		"-v", "error",
		"-threads", "2", // Limita el número de hilos de CPU para reducir drásticamente el consumo de RAM
	}

	// Optimizaciones de hardware para HLS
	args = append(args, getHWAccelArgs()...)

	args = append(args,
		"-ss", strconv.Itoa(startTime),
		"-t", strconv.Itoa(segmentDuration),
		"-i", item.Path,
		"-copyts", // Mantener timestamps para sincronización
	)

	// Configuración inteligente de Video para HLS (Remux vs Transcode)
	if opts.VideoCodec == "copy" {
		args = append(args, "-c:v", "copy")
	} else {
		vCodec := opts.VideoCodec
		if vCodec == "" {
			vCodec = "libx264"
		}
		args = append(args, "-c:v", vCodec, "-preset", "veryfast")

		if opts.MaxHeight > 0 && item.Height > opts.MaxHeight {
			args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", opts.MaxHeight))
		}
	}

	// Configuración inteligente de Audio para HLS
	if opts.AudioCodec == "copy" {
		args = append(args, "-c:a", "copy")
	} else {
		aCodec := opts.AudioCodec
		if aCodec == "" {
			aCodec = "aac"
		}
		args = append(args, "-c:a", aCodec, "-ac", "2", "-b:a", "128k")
	}

	// Formato MPEG-TS para segmentos HLS
	args = append(args, "-f", "mpegts", "-")

	return exec.Command(config.AppConfig.Transcoder.FFmpegPath, args...)
}

func getHWAccelArgs() []string {
	hw := config.AppConfig.Transcoder.HardwareAcceleration
	if hw == "vaapi" {
		return []string{"-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128", "-hwaccel_output_format", "vaapi"}
	} else if hw == "nvenc" {
		return []string{"-hwaccel", "cuda", "-hwaccel_output_format", "cuda"}
	}
	return nil
}
