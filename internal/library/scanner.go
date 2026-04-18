package library

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
)

// WatchFolder inicia el monitoreo de una carpeta de medios en tiempo real.
func WatchFolder(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creando el monitor de archivos: ", err)
	}
	defer watcher.Close()

	// Asegurarse de que la carpeta existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Println("No se pudo crear la carpeta de medios:", path)
			return
		}
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Solo nos interesan creaciones o modificaciones de archivos de video
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					processFile(event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					removeItem(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error del monitor de archivos:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal("Error añadiendo carpeta al monitor: ", err)
	}
	
	// Escaneo inicial rápido
	log.Println("Iniciando escaneo inicial de:", path)
	scanInitial(path)

	log.Printf("Monitor de biblioteca activo en: %s (Bajo consumo de RAM)", path)
	<-done
}

// scanInitial recorre la carpeta una vez al inicio.
func scanInitial(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			processFile(path)
		}
		return nil
	})
}

// processFile añade o actualiza un archivo en la base de datos si es video.
func processFile(path string) {
	if !isVideoFile(path) {
		return
	}

	name := filepath.Base(path)
	ext := filepath.Ext(path)
	
	var item models.MediaItem
	err := database.DB.Where("path = ?", path).First(&item).Error
	
	if err != nil { // No existe, crear uno nuevo
		newItem := models.MediaItem{
			Name:      strings.TrimSuffix(name, ext),
			Path:      path,
			Type:      "Movie", // Por ahora asumimos películas
			Container: strings.TrimPrefix(ext, "."),
		}
		database.DB.Create(&newItem)
		log.Printf("[Library] Nuevo archivo detectado: %s", name)
	}
}

// removeItem elimina un archivo de la base de datos si es borrado del disco.
func removeItem(path string) {
	database.DB.Where("path = ?", path).Delete(&models.MediaItem{})
	log.Printf("[Library] Archivo eliminado: %s", filepath.Base(path))
}

func isVideoFile(path string) bool {
	extensions := []string{".mp4", ".mkv", ".avi", ".mov", ".m4v"}
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range extensions {
		if e == ext {
			return true
		}
	}
	return false
}
