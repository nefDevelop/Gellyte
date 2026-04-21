package library

import (
	"testing"
)

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Archivo MP4 normal", "/media/movies/Iron Man.mp4", true},
		{"Archivo MKV normal", "/media/series/show.mkv", true},
		{"Extensión en mayúsculas", "/media/movies/Thor.MP4", true},
		{"Extensión mixta", "/media/movies/Hulk.MkV", true},
		{"Archivo de subtítulos SRT", "/media/movies/Iron Man.srt", false},
		{"Archivo de imagen JPG", "/media/movies/poster.jpg", false},
		{"Archivo sin extensión", "/media/movies/Iron Man", false},
		{"Ruta con puntos en el nombre", "/media/movies/Mr. Robot.S01E01.mp4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVideoFile(tt.path)
			if result != tt.expected {
				t.Errorf("isVideoFile(%q) = %v; se esperaba %v", tt.path, result, tt.expected)
			}
		})
	}
}
