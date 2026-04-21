package main

import (
	"log"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Gestión de usuarios de Gellyte",
}

var userAddCmd = &cobra.Command{
	Use:   "add [username] [password]",
	Short: "Añade un nuevo usuario",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		isAdmin, _ := cmd.Flags().GetBool("admin")
		runUserAdd(args[0], args[1], isAdmin)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userAddCmd)
	userAddCmd.Flags().Bool("admin", false, "Crear como administrador")
}

func runUserAdd(username, password string, isAdmin bool) {
	config.InitConfig()
	database.InitDB()

	user := models.User{
		Username: username,
		Password: password,
		IsAdmin:  isAdmin,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Fatalf("Error creando usuario: %v", err)
	}

	log.Printf("Usuario '%s' creado correctamente (Admin: %v, ID: %s)", username, isAdmin, user.ID)
}
