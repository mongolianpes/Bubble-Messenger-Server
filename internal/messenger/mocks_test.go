package messenger

import (
	"bubble/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

func NewUsersMockService() *MockUsersService {
	return &MockUsersService{}
}

type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) TLS(ctx context.Context, isRegistring bool, clientPublicKey, id string) (string, error) {
	args := m.Called(ctx, isRegistring, clientPublicKey, id)
	return args.String(0), args.Error(1)
}

func (m *MockUsersService) Register(ctx context.Context, login, name, password, device string) (string, error) {
	args := m.Called(ctx, login, name, password, device)
	return args.String(0), args.Error(1)
}

func (m *MockUsersService) Auth(ctx context.Context, login, password, device string) (string, string, error) {
	args := m.Called(ctx, login, password, device)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUsersService) GetAuthInfo(ctx context.Context, login, password, device string) (string, int, error) {
	args := m.Called(ctx, login, password, device)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *MockUsersService) Search(ctx context.Context, login string) ([]models.FindUser, error) {
	args := m.Called(ctx, login)
	return args.Get(0).([]models.FindUser), args.Error(1)
}

func (m *MockUsersService) GetInfoByLogin(ctx context.Context, login string) (models.FindUser, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(models.FindUser), args.Error(1)
}

func (m *MockUsersService) GetInfoByID(ctx context.Context, id int) (models.FindUser, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.FindUser), args.Error(1)
}

func (m *MockUsersService) SetAvatar(ctx context.Context, login, avatarPath string) error {
	args := m.Called(ctx, login, avatarPath)
	return args.Error(0)
}

func (m *MockUsersService) GetAvatar(ctx context.Context, login string) (string, error) {
	args := m.Called(ctx, login)
	return args.String(0), args.Error(1)
}

func (m *MockUsersService) Close() error {
	args := m.Called()
	return args.Error(0)
}
