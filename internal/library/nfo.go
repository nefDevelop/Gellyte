package library

import (
	"encoding/xml"
	"os"
)

// MovieNfo representa la estructura básica de un archivo .nfo de película
type MovieNfo struct {
	XMLName xml.Name `xml:"movie"`
	Title   string   `xml:"title"`
	Year    int      `xml:"year"`
	Plot    string   `xml:"plot"`
}

// ParseMovieNfo lee y decodifica un archivo .nfo
func ParseMovieNfo(path string) (*MovieNfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nfo MovieNfo
	if err := xml.Unmarshal(data, &nfo); err != nil {
		return nil, err
	}

	return &nfo, nil
}
