package services

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

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

	// Generar un token único de sesión (32 chars hex sin guiones)
	hash := md5.Sum([]byte(time.Now().String() + user.Username + deviceID))
	token := hex.EncodeToString(hash[:])

	return user, token, nil
}

func (s *authService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *authService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.ListAll()
}
