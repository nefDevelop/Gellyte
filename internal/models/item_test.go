package models

import (
	"testing"
)

func TestMediaItem_BeforeCreate(t *testing.T) {
	item1 := MediaItem{Path: "/media/movies/Iron Man.mp4"}
	item2 := MediaItem{Path: "/media/movies/Iron Man.mp4"}
	item3 := MediaItem{Path: "/media/movies/Thor.mp4"}

	_ = item1.BeforeCreate(nil)
	_ = item2.BeforeCreate(nil)
	_ = item3.BeforeCreate(nil)

	if len(item1.ID) != 32 {
		t.Errorf("Se esperaba una longitud de ID de 32 caracteres, se obtuvo %d", len(item1.ID))
	}

	if item1.ID != item2.ID {
		t.Errorf("Se esperaban IDs idénticos para la misma ruta. Se obtuvo %s y %s", item1.ID, item2.ID)
	}

	if item1.ID == item3.ID {
		t.Errorf("Se esperaban IDs diferentes para rutas diferentes. Ambos obtuvieron %s", item1.ID)
	}
}
