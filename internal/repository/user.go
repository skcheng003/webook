package repository

import (
	"context"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/repository/dao"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNoFound = dao.ErrUserNoFound

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Nickname: u.Nickname,
		Birth:    u.Birth,
		Bio:      u.Bio,
	}, err
}

func (r *UserRepository) FindByUid(ctx context.Context, uid int64) (domain.User, error) {
	u, err := r.dao.FindByUid(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Nickname: u.Nickname,
		Birth:    u.Birth,
		Bio:      u.Bio,
	}, err
}

func (r *UserRepository) EditProfile(ctx context.Context, user domain.User) error {
	_, err := r.dao.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	return r.dao.EditProfile(ctx, dao.User{
		Email:    user.Email,
		Nickname: user.Nickname,
		Birth:    user.Birth,
		Bio:      user.Bio,
	})
}
