package service

import (
	"TestSeparation/internal/entity"
	"TestSeparation/internal/pkg/validation"
	"TestSeparation/internal/repository"
	"context"
	"errors"
)

var (
	ErrInValidInput      = errors.New("invalid input")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UsersService struct {
	usersRepository repository.UsersRepository
}

func NewUsersService(usersRepository repository.UsersRepository) *UsersService {
	return &UsersService{usersRepository}
}

// Signup creates a new user if the phone number is not already registered
// todo: use unique index on phone number
func (u *UsersService) Signup(ctx context.Context, phoneNumber string, displayName string) (entity.User, error) {

	if !validation.IsValidPhoneNumber(phoneNumber) {
		return entity.User{}, ErrInValidInput
	}

	_, err := u.usersRepository.ByField(ctx, "phone_number", phoneNumber)

	if err == nil {
		return entity.User{}, ErrUserAlreadyExists
	}

	// we must get ErrNotFound error, so if err == nil it means user already exists
	if !errors.Is(err, repository.ErrNotFound) {
		return entity.User{}, err
	}

	// user with this phone number does not exist, so we can create a new user
	user := &entity.User{
		DisplayName: displayName,
		PhoneNumber: phoneNumber,
	}

	if err := u.usersRepository.Create(ctx, user); err != nil {
		return entity.User{}, err
	}

	return *user, nil
}
