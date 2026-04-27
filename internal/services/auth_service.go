package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Authenticate(username, password, deviceID string) (*models.User, string, error)
	GetUserByID(id string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Authenticate(username, password, deviceID string) (*models.User, string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", errors.New("usuario no encontrado")
	}

	// Verificar password usando bcrypt
	if user.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			return nil, "", errors.New("password incorrecta")
		}
	}

	// Generar un token criptográficamente seguro de 32 caracteres (16 bytes = 32 hex chars)
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	return user, token, nil
}

func (s *authService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *authService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.ListAll()
}
