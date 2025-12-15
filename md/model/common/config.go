package common

var (
	Port             string // 端口
	LogPath          string // 日志目录
	DataPath         string // 数据目录
	Register         bool   // 允许注册
	BasicTokenKey    string // token相关接口认证key前缀
	ResourceName     string // 静态资源目录名，在数据目录下
	PictureName      string // 图片目录名，在静态资源目录下
	ThumbnailName    string // 缩略图目录名，在静态资源目录下
	PostgresHost     string // postgres主机地址
	PostgresPort     string // postgres端口
	PostgresUser     string // postgres用户
	PostgresPassword string // postgres密码
	PostgresDB       string // postgres数据库名
	AIEncryptKey     string // AI API Key加密密钥
)
