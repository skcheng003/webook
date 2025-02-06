package ioc

import (
	"github.com/skcheng003/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := viper.GetString("db.mysql.dsn")
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		// TODO: use logger replace panic
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
