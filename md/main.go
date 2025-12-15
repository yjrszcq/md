package main

import (
	"embed"
	"flag"
	"io/fs"
	"md/controller"
	"md/middleware"
	"md/model/common"
	"md/util"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
)

//go:embed web
var web embed.FS

func init() {
	// 解析命令行参数
	flag.StringVar(&common.Port, "p", "9900", "监听端口")
	flag.StringVar(&common.LogPath, "log", "./logs", "日志目录，存放近30天的日志，设置为空则不生成日志文件")
	flag.StringVar(&common.DataPath, "data", "./data", "数据目录，存放数据库文件和图片")
	flag.BoolVar(&common.Register, "reg", true, "是否允许注册（即使禁止注册，在没有任何用户的情况时仍可注册）")
	flag.StringVar(&common.PostgresHost, "pg_host", "", "postgres主机地址")
	flag.StringVar(&common.PostgresPort, "pg_port", "", "postgres端口")
	flag.StringVar(&common.PostgresUser, "pg_user", "", "postgres用户")
	flag.StringVar(&common.PostgresPassword, "pg_password", "", "postgres密码")
	flag.StringVar(&common.PostgresDB, "pg_db", "", "postgres数据库名")
	flag.StringVar(&common.AIEncryptKey, "ai_key", "md-ai-encrypt-key-2024", "AI API Key加密密钥")
	flag.Parse()

	// 固定配置
	common.DataPath = util.PathCompletion(common.DataPath)
	common.BasicTokenKey = "md"
	common.ResourceName = "resource"
	common.PictureName = "picture"
	common.ThumbnailName = "thumbnail"
}

func main() {
	// 创建iris服务
	app := iris.New()

	// 初始化日志
	middleware.InitLog(app.Logger())

	// 全局异常恢复
	app.Use(middleware.GlobalRecover)

	// gzip压缩
	app.Use(iris.Compression)

	// 初始化雪花算法节点
	err := util.InitSnowflake(0)
	if err != nil {
		middleware.Log.Error("初始化雪花算法节点失败：", err)
		return
	}

	// 初始化数据目录
	err = middleware.InitDataDir(common.DataPath, common.ResourceName, common.PictureName, common.ThumbnailName)
	if err != nil {
		return
	}

	// 初始化数据库连接
	err = middleware.InitDB()
	if err != nil {
		return
	}

	// 初始化API路由
	controller.InitRouter(app)

	// 网页资源路由
	app.Use(iris.StaticCache(time.Hour * 720))
	webFs, err := fs.Sub(web, "web")
	if err != nil {
		middleware.Log.Error("初始化网页资源失败：", err)
		return
	}
	app.HandleDir("/", http.FS(webFs))

	// 静态资源路由
	app.HandleDir("/"+common.ResourceName, common.DataPath+common.ResourceName)

	// 启动服务
	app.Logger().Error(app.Run(iris.Addr(":" + common.Port)))
}
