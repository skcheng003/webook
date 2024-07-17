package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("email address conflict")
	ErrUserNoFound   = gorm.ErrRecordNotFound
)

type UserDao interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByUid(ctx context.Context, uid int64) (User, error)
	Insert(ctx context.Context, u User) error
	EditProfile(ctx context.Context, u User) error
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDao {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByUid(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("Id = ?", uid).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli() // 存毫秒数
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GORMUserDAO) EditProfile(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Where("Id = ?", u.Id).
		Updates(User{Nickname: u.Nickname, Birth: u.Birth, Bio: u.Bio}).Error
	return err
}

// User 直接对应数据库表结构，entity 或 Model
// PO(persistent object)
type User struct {
	Id int64 `gorm:"primaryKey, autoIncrement"`
	// 唯一索引允许有多个NULL，但不可以有多个“”
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	Nickname string `gorm:"size: 16"`
	Birth    string
	Bio      string `gorm:"size: 256"`
	Ctime    int64
	Utime    int64
}
