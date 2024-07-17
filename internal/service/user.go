package service

import (
	"context"
	"errors"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("invalid user or password")
var ErrUserNoFound = repository.ErrUserNoFound

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	EditProfile(ctx context.Context, user domain.User) error
	FindProfile(ctx context.Context, email string) (domain.User, error)
	FindProfileJWT(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.CreateUser(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
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

func (svc *userService) EditProfile(ctx context.Context, user domain.User) error {
	return svc.repo.EditProfile(ctx, user)
}

func (svc *userService) FindProfile(ctx context.Context, email string) (domain.User, error) {
	return svc.repo.FindByEmail(ctx, email)
	// return domain.User{}, nil
}

func (svc *userService) FindProfileJWT(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindByUid(ctx, uid)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNoFound) {
		// err == nil 和 err != ErrUserNotFound 都会进入这个分支
		// 快路径
		return u, err
	}
	// 未找到用户，创建用户
	// 慢路径，在系统资源不足，触发降级之后，不能走慢路径了
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.CreateUser(ctx, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 存在主从延迟问题
	return svc.repo.FindByPhone(ctx, phone)
}
