package library

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
)

// OnLibraryChanged is called when a change is detected in the library.
var OnLibraryChanged func()

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
		log.Printf("[Scanner] La carpeta no existe, creándola: %s", path)
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
							watcher.Add(event.Name) // Monitorizar la subcarpeta nueva en tiempo real
							enqueueTask(event.Name, libType, path, true)
							// Escaneo recursivo inicial de la nueva carpeta
							scanInitial(event.Name, libType, path)
						} else {
							enqueueTask(event.Name, libType, path, false)
						}
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					removeItem(event.Name, path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error del monitor de archivos:", err)
			}
		}
	}()

	// Añadir la carpeta principal y todas sus subcarpetas existentes al monitor
	filepath.WalkDir(path, func(walkPath string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			watcher.Add(walkPath)
		}
		return nil
	})

	// Escaneo inicial rápido
	log.Println("Iniciando escaneo inicial de:", path)
	scanInitial(path, libType, path)

	//log.Printf("Monitor de biblioteca activo en: %s (Tipo: %s)", path, libType)
	<-done
}

func ScanManual(path string, libType string) {
	path = filepath.Clean(path)
	log.Println("Escaneando:", path)
	scanInitial(path, libType, path)
}

var (
	workerWg  sync.WaitGroup
	taskQueue = make(chan scanTask, 500)
)

type scanTask struct {
	path    string
	libType string
	libRoot string
	isDir   bool
}

func init() {
	// Determinar el número de workers dinámicamente según la CPU para acelerar ffprobe
	workers := runtime.NumCPU() * 2
	if workers > 16 {
		workers = 16 // Límite máximo para no saturar la RAM con demasiados procesos de ffprobe
	}

	for i := 0; i < workers; i++ {
		go func() {
			for task := range taskQueue {
				if task.isDir {
					processDirectory(task.path, task.libType, task.libRoot)
				} else {
					processFile(task.path, task.libType)
				}
				workerWg.Done()
			}
		}()
	}
}

// StopScanner espera a que terminen las tareas en curso.
func StopScanner() {
	workerWg.Wait()
	log.Println("[Scanner] Tareas finalizadas correctamente.")
}

func enqueueTask(path string, libType string, libRoot string, isDir bool) {
	workerWg.Add(1)
	taskQueue <- scanTask{
		path:    path,
		libType: libType,
		libRoot: libRoot,
		isDir:   isDir,
	}
}

// scanInitial recorre la carpeta una vez al inicio utilizando WalkDir (más rápido).
func scanInitial(root string, libType string, libRoot string) {
	log.Printf("[Scanner] Iniciando WalkDir en: %s", root)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("[Scanner] Error accediendo a ruta %s: %v", path, err)
			return err
		}
		// Log temporal para depuración
		// log.Printf("[Scanner] Visitando: %s (Es carpeta: %v)", path, d.IsDir())

		enqueueTask(path, libType, libRoot, d.IsDir())
		return nil
	})
	if err != nil {
		log.Printf("[Scanner] Error en el escaneo inicial de %s: %v", root, err)
	}
}

// processFile añade o actualiza un archivo en la base de datos si es video.
func processFile(path string, libType string) {
	path = filepath.Clean(path)
	if !isVideoFile(path) {
		return
	}

	log.Printf("[Scanner] Procesando archivo: %s", filepath.Base(path))

	name := filepath.Base(path)
	ext := filepath.Ext(path)

	var item models.MediaItem
	err := database.DB.Where("path = ?", path).First(&item).Error
	if err == nil {
		// El archivo ya existe. Si no tiene metadatos, intentamos extraerlos ahora.
		if item.Width == 0 {
			meta, err := GetVideoMetadata(path)
			if err == nil {
				item.RunTimeTicks = meta.DurationTicks
				item.Width = meta.Width
				item.Height = meta.Height
				item.Bitrate = meta.Bitrate
				item.VideoCodec = meta.VideoCodec
				item.AudioCodec = meta.AudioCodec
				item.MediaStreams = meta.Streams
				database.DB.Save(&item)
				log.Printf("[Scanner] ! Metadatos actualizados para: %s", name)
			}
		}
		return
	}

	// No existe, crear uno nuevo
	itemType := "Movie"
	parentId := strings.ReplaceAll(config.AppConfig.Jellyfin.MoviesLibraryID, "-", "")

	if libType == "series" {
		itemType = "Episode"
		// Intentar encontrar el ParentId (Season o Series)
		parentPath := filepath.Dir(path)
		var parent models.MediaItem
		if err := database.DB.Where("path = ?", parentPath).First(&parent).Error; err == nil {
			parentId = parent.ID
		} else {
			// Si no hay carpeta padre en DB, usamos la biblioteca de Series por defecto
			parentId = strings.ReplaceAll(config.AppConfig.Jellyfin.SeriesLibraryID, "-", "")
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
		newItem.AudioCodec = meta.AudioCodec
		newItem.MediaStreams = meta.Streams
	} else {
		log.Printf("[Scanner] Advertencia: No se extrajeron metadatos de '%s'. ¿Está ffprobe instalado? Error: %v", name, err)
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
	log.Printf("[Scanner] + Añadido a biblioteca: %s (%s)", newItem.Name, newItem.Type)
	if OnLibraryChanged != nil {
		OnLibraryChanged()
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
			parentId = strings.ReplaceAll(config.AppConfig.Jellyfin.SeriesLibraryID, "-", "")
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
	log.Printf("[Scanner] + Carpeta detectada: %s (%s)", name, itemType)
	if OnLibraryChanged != nil {
		OnLibraryChanged()
	}
}

// removeItem elimina un archivo de la base de datos si es borrado del disco.
func removeItem(itemPath string, libRoot string) {
	// Protección contra desconexión de disco duro:
	// Verificamos si la raíz de la biblioteca sigue accesible.
	// Si el disco se desconextó, la raíz dará error y evitamos borrar la base de datos.
	if _, err := os.Stat(libRoot); os.IsNotExist(err) {
		return
	}

	database.DB.Where("path = ?", itemPath).Delete(&models.MediaItem{})
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
