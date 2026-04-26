package main

import (
	"log"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
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

var userDeleteCmd = &cobra.Command{
	Use:   "delete [username]",
	Short: "Elimina un usuario",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runUserDelete(args[0])
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userDeleteCmd)
	userAddCmd.Flags().Bool("admin", false, "Crear como administrador")
}

func runUserAdd(username, password string, isAdmin bool) {
	config.InitConfig()
	database.InitDB()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hasheando contraseña: %v", err)
	}

	userID := uuid.New().String()

	user := models.User{
		ID:       userID,
		Username: username,
		Password: string(hashedPassword),
		IsAdmin:  isAdmin,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Fatalf("Error creando usuario: %v", err)
	}

	log.Printf("Usuario '%s' creado correctamente (Admin: %v, ID: %s)", username, isAdmin, user.ID)
}

func runUserDelete(username string) {
	config.InitConfig()
	database.InitDB()

	if err := database.DB.Where("username = ?", username).Delete(&models.User{}).Error; err != nil {
		log.Fatalf("Error eliminando usuario: %v", err)
	}

	log.Printf("Usuario '%s' eliminado correctamente", username)
}
