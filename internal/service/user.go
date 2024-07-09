package service

import (
	"context"
	"errors"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("invalid user or password")
var ErrUserNoFound = repository.ErrUserNoFound

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.CreateUser(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, ErrUserNoFound) {
		return domain.User{}, ErrUserNoFound
	}

	if err != nil {
		// TODO: need more actions
		return domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// TODO: add info logger here
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) EditProfile(ctx context.Context, user domain.User) error {
	return svc.repo.EditProfile(ctx, user)
}

func (svc *UserService) FindProfile(ctx context.Context, email string) (domain.User, error) {
	return svc.repo.FindByEmail(ctx, email)
	// return domain.User{}, nil
}

func (svc *UserService) FindProfileJWT(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindByUid(ctx, uid)
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	svc.repo.FindByPhone(ctx, phone)
	return domain.User{}, nil
}
