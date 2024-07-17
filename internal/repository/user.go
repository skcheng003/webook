package repository

import (
	"context"
	"database/sql"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/repository/cache"
	"github.com/skcheng003/webook/internal/repository/dao"
	"time"
)

var ErrUserDuplicate = dao.ErrUserDuplicate
var ErrUserNoFound = dao.ErrUserNoFound

type UserRepository interface {
	CreateUser(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByUid(ctx context.Context, uid int64) (domain.User, error)
	EditProfile(ctx context.Context, user domain.User) error
}

type userRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &userRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err == nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), err
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), err
}

func (r *userRepository) FindByUid(ctx context.Context, uid int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, uid)
	if err == nil {
		// cache hit, 直接返回
		return u, err
	}
	// cache miss, 从数据库中读取
	// 在 orm 层用中间件对数据库访问进行限流，防止数据库被打爆
	// TODO: add cache miss log
	ue, err := r.dao.FindByUid(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(ue)
	go func() {
		// 异步写入缓存
		err = r.cache.Set(ctx, u)
		if err != nil {
			// TODO: add some log, set cache failed
		}
	}()
	return u, err
}

func (r *userRepository) EditProfile(ctx context.Context, user domain.User) error {
	_, err := r.dao.FindByUid(ctx, user.Id)
	if err != nil {
		return err
	}

	err = r.dao.EditProfile(ctx, dao.User{
		Id:       user.Id,
		Nickname: user.Nickname,
		Birth:    user.Birth,
		Bio:      user.Bio,
	})
	return err
}

func (r *userRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Ctime:    user.Ctime.UnixMilli(),
	}
}

func (r *userRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Password: user.Password,
		Ctime:    time.UnixMilli(user.Ctime),
	}
}
