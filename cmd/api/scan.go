package main

import (
	"log"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/library"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Ejecuta un escaneo completo de la biblioteca de medios",
	Run: func(cmd *cobra.Command, args []string) {
		runScan()
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan() {
	config.InitConfig()
	database.InitDB()

	log.Println("[Scanner] Iniciando escaneo manual...")
	
	// Realizar escaneo manual (bloqueante para CLI)
	// Como WatchFolder es infinito, llamamos a scanInitial directamente si estuviera expuesta,
	// o simplemente logueamos que el escaneo se realiza al iniciar el monitor (pero aquí queremos que termine).
	
	// Refactorizamos library para exponer un escaneo manual si es necesario.
	// Por ahora, usamos el comportamiento de WatchFolder pero sin el loop infinito si podemos.
	
	library.ScanManual(config.AppConfig.Library.MoviesPath, "movies")
	library.ScanManual(config.AppConfig.Library.SeriesPath, "series")
	
	log.Println("[Scanner] Escaneo completado.")
}
