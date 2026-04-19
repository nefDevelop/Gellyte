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
func WatchFolder(path string, libType string) {
	path = filepath.Clean(path)
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
				// Crear/Modificar
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					info, err := os.Stat(event.Name)
					if err == nil {
						if info.IsDir() {
							processDirectory(event.Name, libType, path)
							// Escaneo recursivo inicial de la nueva carpeta
							scanInitial(event.Name, libType, path)
						} else {
							processFile(event.Name, libType, path)
						}
					}
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
	scanInitial(path, libType, path)

	log.Printf("Monitor de biblioteca activo en: %s (Tipo: %s)", path, libType)
	<-done
}

// scanInitial recorre la carpeta una vez al inicio.
func scanInitial(root string, libType string, libRoot string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if info.IsDir() {
				processDirectory(path, libType, libRoot)
			} else {
				processFile(path, libType, libRoot)
			}
		}
		return nil
	})
}

// processFile añade o actualiza un archivo en la base de datos si es video.
func processFile(path string, libType string, libRoot string) {
	path = filepath.Clean(path)
	if !isVideoFile(path) {
		return
	}

	name := filepath.Base(path)
	ext := filepath.Ext(path)
	
	var item models.MediaItem
	err := database.DB.Where("path = ?", path).First(&item).Error
	
	if err != nil { // No existe, crear uno nuevo
		itemType := "Movie"
		parentId := "12345678-1234-1234-1234-123456789012" // Default Movies Library

		if libType == "series" {
			itemType = "Episode"
			// Intentar encontrar el ParentId (Season o Series)
			parentPath := filepath.Dir(path)
			var parent models.MediaItem
			if err := database.DB.Where("path = ?", parentPath).First(&parent).Error; err == nil {
				parentId = parent.ID
			} else {
				// Si no hay carpeta padre en DB, usamos la biblioteca de Series por defecto
				parentId = "22345678-1234-1234-1234-123456789012"
			}
		}

		newItem := models.MediaItem{
			Name:      strings.TrimSuffix(name, ext),
			Path:      path,
			Type:      itemType,
			ParentID:  parentId,
			Container: strings.TrimPrefix(ext, "."),
		}

		// Extraer metadatos técnicos con ffprobe
		meta, err := GetVideoMetadata(path)
		if err == nil {
			newItem.RunTimeTicks = meta.DurationTicks
			newItem.Width = meta.Width
			newItem.Height = meta.Height
			newItem.Bitrate = meta.Bitrate
			newItem.VideoCodec = meta.VideoCodec
		}

		// Buscar metadatos en archivo .nfo
		nfoPath := strings.TrimSuffix(path, ext) + ".nfo"
		if nfo, err := ParseMovieNfo(nfoPath); err == nil {
			newItem.Name = nfo.Title
			newItem.ProductionYear = nfo.Year
			newItem.Overview = nfo.Plot
		}

		// Generar miniatura si no existe una imagen local
		dir := filepath.Dir(path)
		thumbPath := filepath.Join(dir, "thumb.jpg")
		if _, err := os.Stat(thumbPath); os.IsNotExist(err) && newItem.Width > 0 {
			GenerateThumbnail(path, thumbPath)
		}

		database.DB.Create(&newItem)
	}
}

// processDirectory maneja la creación de carpetas (Series/Seasons)
func processDirectory(path string, libType string, libRoot string) {
	path = filepath.Clean(path)
	if path == libRoot {
		return
	}

	var item models.MediaItem
	err := database.DB.Where("path = ?", path).First(&item).Error
	if err == nil {
		return // Ya existe
	}

	name := filepath.Base(path)
	itemType := "Folder"
	parentId := ""

	if libType == "series" {
		parentPath := filepath.Dir(path)
		if parentPath == libRoot || parentPath == "." {
			itemType = "Series"
			parentId = "22345678-1234-1234-1234-123456789012"
		} else {
			itemType = "Season"
			// Buscar la Serie padre
			var seriesParent models.MediaItem
			if err := database.DB.Where("path = ?", parentPath).First(&seriesParent).Error; err == nil {
				parentId = seriesParent.ID
			}
		}
	}

	newItem := models.MediaItem{
		Name:     name,
		Path:     path,
		Type:     itemType,
		ParentID: parentId,
	}

	database.DB.Create(&newItem)
	log.Printf("[Library] Carpeta detectada: %s (%s)", name, itemType)
}

// removeItem elimina un archivo de la base de datos si es borrado del disco.
func removeItem(path string) {
	database.DB.Where("path = ?", path).Delete(&models.MediaItem{})
	log.Printf("[Library] Item eliminado: %s", filepath.Base(path))
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
