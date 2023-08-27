package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("email address conflict")
	ErrUserNoFound        = gorm.ErrRecordNotFound
)

const (
	uniqueConflictsErrNo uint16 = 1062
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) FindByUid(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("Id = ?", uid).First(&u).Error
	return u, err
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli() // 存毫秒数
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) EditProfile(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Where("Id = ?", u.Id).
		Updates(User{Nickname: u.Nickname, Birth: u.Birth, Bio: u.Bio}).Error
	return err
}

// User 直接对应数据库表结构，entity 或 Model
// PO(persistent object)
type User struct {
	Id       int64  `gorm:"primaryKey, autoIncrement"`
	Email    string `gorm:"unique"`
	Password string
	Nickname string `gorm:"size: 16"`
	Birth    string
	Bio      string `gorm:"size: 256"`
	Ctime    int64
	Utime    int64
}
