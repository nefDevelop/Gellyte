package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Name string `mapstructure:"name"`
	} `mapstructure:"server"`

	Database struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	Jellyfin struct {
		ServerUUID      string `mapstructure:"server_uuid"`
		AdminUUID       string `mapstructure:"admin_uuid"`
		MoviesLibraryID string `mapstructure:"movies_library_id"`
		SeriesLibraryID string `mapstructure:"series_library_id"`
	} `mapstructure:"jellyfin"`

	Transcoder struct {
		FFmpegPath  string `mapstructure:"ffmpeg_path"`
		FFprobePath string `mapstructure:"ffprobe_path"`
	} `mapstructure:"transcoder"`

	Library struct {
		MoviesPath string `mapstructure:"movies_path"`
		SeriesPath string `mapstructure:"series_path"`
	} `mapstructure:"library"`
}

var AppConfig Config

func InitConfig() {
	// Cargar .env si existe
	err := godotenv.Load()
	if err != nil {
		log.Println("[Config] No se encontró archivo .env, usando variables de entorno o valores por defecto.")
	}

	// Configurar Viper
	viper.SetConfigName("config") // nombre del archivo (sin extensión)
	viper.SetConfigType("yaml")   // o "json", "toml", etc.
	viper.AddConfigPath(".")      // buscar en el directorio actual
	viper.AddConfigPath("./config")

	// Valores por defecto
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.name", "Gellyte")
	viper.SetDefault("database.path", "gellyte.db")
	viper.SetDefault("jellyfin.server_uuid", "83e4c49d-9273-4556-9a5d-4952011702f3")
	viper.SetDefault("jellyfin.admin_uuid", "53896590-3b41-46a4-9591-96b054a8e3f6")
	viper.SetDefault("jellyfin.movies_library_id", "12345678-1234-1234-1234-123456789012")
	viper.SetDefault("jellyfin.series_library_id", "22345678-1234-1234-1234-123456789012")
	viper.SetDefault("transcoder.ffmpeg_path", "/usr/bin/ffmpeg")
	viper.SetDefault("transcoder.ffprobe_path", "/usr/bin/ffprobe")
	viper.SetDefault("library.movies_path", "./media/peliculas")
	viper.SetDefault("library.series_path", "./media/series")

	// Leer variables de entorno (ej: GELLYTE_SERVER_PORT)
	viper.SetEnvPrefix("GELLYTE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Leer archivo de configuración
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("[Config] Archivo config.yaml no encontrado, usando valores por defecto.")
		} else {
			log.Fatalf("[Config] Error leyendo archivo de configuración: %v", err)
		}
	}

	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("[Config] No se pudo deserializar la configuración: %v", err)
	}

	log.Println("[Config] Configuración cargada correctamente.")
}
