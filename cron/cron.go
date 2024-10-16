package conf

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"qa/cache"
	"qa/model"
)

// Init 初始化配置项
func Init() {
	// 从本地读取环境变量
	_ = godotenv.Load()

	// 设置运行模式
	gin.SetMode(os.Getenv("GIN_MODE"))

	// 启动各种连接单例
	model.Database(os.Getenv("MYSQL_DSN"))
	cache.Redis()
}
