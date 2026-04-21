package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gellyte",
	Short: "Gellyte: Un servidor compatible con Jellyfin escrito en Go",
	Long: `Gellyte es un servidor de medios ultra-ligero y rápido que implementa 
la API de Jellyfin para permitir la conexión de aplicaciones oficiales.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Aquí se pueden añadir banderas globales si se desea
}
