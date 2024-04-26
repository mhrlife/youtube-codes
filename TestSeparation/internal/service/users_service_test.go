package service

import (
	"TestSeparation/internal/entity"
	"TestSeparation/internal/repository"
	"TestSeparation/internal/repository/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Signup_UserNotExists(t *testing.T) {

	usersRepository := &mocks.UsersRepository{}

	usersRepository.On("ByField", mock.Anything, "PhoneNumber", "09211231231").
		Return(entity.User{}, repository.ErrNotFound).Once()

	usersRepository.On("Create", mock.Anything, mock.MatchedBy(func(user *entity.User) bool {
		return user.PhoneNumber == "09211231231" && user.DisplayName == "Mohammad"
	})).Return(nil).Once()

	service := NewUsersService(usersRepository)
	_, err := service.Signup(context.Background(), "09211231231", "Mohammad")

	assert.NoError(t, err)

	usersRepository.AssertExpectations(t)

}

func TestUserService_Signup_UserExists(t *testing.T) {

	usersRepository := &mocks.UsersRepository{}

	usersRepository.On("ByField", mock.Anything, "phone_number", "09211231231").
		Return(entity.User{}, nil).Once()

	service := NewUsersService(usersRepository)
	_, err := service.Signup(context.Background(), "09211231231", "Mohammad")

	assert.ErrorIs(t, err, ErrUserAlreadyExists)

	usersRepository.AssertExpectations(t)

}
