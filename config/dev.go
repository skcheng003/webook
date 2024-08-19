//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:739662@tcp(localhost:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
