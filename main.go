package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initViper()
	initLogger()

	server := initWebServer()
	zap.L().Info("开始监听8081端口")
	server.Run(":8081")
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("%v", err)
		panic("read config error")
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
