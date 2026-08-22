package cli

import (
	"errors"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 这里将在 run 之后执行
func loadConfig() {
	explicit := viper.GetString("config") != ""
	if explicit {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err == nil {
		return
	}
	// P1 修复：原实现吞掉全部错误。搜索模式下「未找到配置文件」属正常（纯 flag/环境变量启动）；
	// 但显式指定路径（-c）缺失、或 YAML 解析失败必须报错——
	// 否则服务会带着默认配置静默起跑（叠加空 JWT 密钥等问题，隐患极大）。
	var cfgNotFound viper.ConfigFileNotFoundError
	if !explicit && errors.As(err, &cfgNotFound) {
		return
	}
	panic(exception.New("配置文件加载失败: " + err.Error()))
}

func bindDefaultFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("config", "c", "", "配置文件全路径")
	cmd.PersistentFlags().String(configkey.ProjectDir, ".", "项目目录")
	cmd.PersistentFlags().String(configkey.ProjectName, "app", "项目名称")
	cmd.PersistentFlags().String(configkey.ProjectSubDir4PublicDownload, "", "项目目录中用于公共下载的开放目录（一层），逗号分隔，.表示所有")
	cmd.PersistentFlags().String(configkey.ProjectSubDir4PrivateDownload, "", "项目目录中用于私有下载的开放目录（一层），逗号分隔，.表示所有")
	cmd.PersistentFlags().Bool(configkey.ProfileDev, false, "开发模式 default:false")
	cmd.PersistentFlags().String(configkey.TimeLocation, "Asia/Shanghai", "项目中用到的时区")

	cmd.PersistentFlags().Int(configkey.CacheWrapperTTL, 1, "wrapper ttl 默认1s")

	cmd.PersistentFlags().String(configkey.RedisPrefix, "", "redis key的前缀")
	cmd.PersistentFlags().String(configkey.RedisHost, "", "redis host")
	cmd.PersistentFlags().String(configkey.RedisPort, "", "")
	cmd.PersistentFlags().String(configkey.RedisDB, "", "redis db 数据库号")
	cmd.PersistentFlags().String(configkey.RedisPwd, "", "")

	cmd.PersistentFlags().String(configkey.LogPath, "", "日志目录；空则表示在project.dir/log下；不填不开启文件日志")
	cmd.PersistentFlags().String(configkey.LogName, "main", "日志文件名，无后缀")
	cmd.PersistentFlags().Int(configkey.LogMaxRemain, 0, "最大保留天数")
	cmd.PersistentFlags().Int(configkey.LogMaxBackups, 0, "最大保留个数")
	cmd.PersistentFlags().Int(configkey.LogMaxSize, 20, "单文件最大尺寸")
	cmd.PersistentFlags().String(configkey.LogLevel, "", "日志等级 debug/info/warn/error")
	cmd.PersistentFlags().String(configkey.LogType, "text", "日志写入时的格式 text/json")

	cmd.PersistentFlags().String(configkey.RestServerBase, "", "rest base url")
	cmd.PersistentFlags().String(configkey.RestServerPort, "10000", "")
	cmd.PersistentFlags().String(configkey.RestRequestBodySize, "", "限制request最大，单位MB")
	cmd.PersistentFlags().Bool(configkey.RestPPROF, false, "开启pprof, /debug/pprof")

	cmd.PersistentFlags().Int(configkey.RestReadTimeout, 60, "读完整请求(含body)超时/秒，0不限制")
	cmd.PersistentFlags().Int(configkey.RestReadHeaderTimeout, 10, "读请求头超时/秒(防slowloris慢速攻击)，0不限制")
	cmd.PersistentFlags().Int(configkey.RestWriteTimeout, 0, "响应写超时/秒；0不限制(SSE等长连接场景必须保持0)")
	cmd.PersistentFlags().Int(configkey.RestIdleTimeout, 120, "keep-alive空闲连接回收/秒，0不限制")

	cmd.PersistentFlags().Int(configkey.JwtExpire, 6, "jwt 过期时间/小时")
	cmd.PersistentFlags().String(configkey.JwtSecretKey, "", "jwt 密钥（必填；为空时签发/解析 token 将直接报错，禁止使用可预测的默认密钥）")

	cmd.PersistentFlags().String(configkey.DBDriver, "", "postgres/mysql/mssql")
	cmd.PersistentFlags().String(configkey.DBHost, "", "")
	cmd.PersistentFlags().String(configkey.DBPort, "", "")
	cmd.PersistentFlags().String(configkey.DBName, "", "")
	cmd.PersistentFlags().String(configkey.DBUser, "", "")
	cmd.PersistentFlags().String(configkey.DBPwd, "", "")
	cmd.PersistentFlags().Int(configkey.DBMaxOpen, 25, "最大连接")
	cmd.PersistentFlags().Int(configkey.DBMaxIdle, 5, "最大空闲连接")
	cmd.PersistentFlags().Int(configkey.DBMaxLife, 10, "单位/分钟")

	cmd.PersistentFlags().String(configkey.OpenApiDescription, "openapi doc", "")
	cmd.PersistentFlags().String(configkey.OpenApiTitle, "openapi doc", "")
	cmd.PersistentFlags().String(configkey.OpenApiVersion, "1.0.0", "")
	cmd.PersistentFlags().String(configkey.OpenApiContactName, "", "")
	cmd.PersistentFlags().String(configkey.OpenApiContactUrl, "", "")
	cmd.PersistentFlags().String(configkey.OpenApiContactEmail, "", "")

	cmd.PersistentFlags().String(configkey.AmapKey, "", "高德key")

	cmd.PersistentFlags().String(configkey.AliRegionId, "cn-hangzhou", "ali")
	cmd.PersistentFlags().String(configkey.AliAccessKey, "", "ali")
	cmd.PersistentFlags().String(configkey.AliAccessKeySecret, "", "ali")
	cmd.PersistentFlags().String(configkey.AliSMSTemplate1, "", "ali sms 模板1")
	cmd.PersistentFlags().String(configkey.AliSMSSign1, "", "ali sms 签名1")
	cmd.PersistentFlags().String(configkey.AliSTSRoleArn, "", "ali sts")
	cmd.PersistentFlags().String(configkey.AliOSSBucketName, "", "ali oss default bucket")

	// mqtt
	cmd.PersistentFlags().String(configkey.MQTTBroker, "", "eg: tcp://xx.xx.xx")
	cmd.PersistentFlags().String(configkey.MQTTClientID, "client", "")
	cmd.PersistentFlags().String(configkey.MQTTUsername, "", "")
	cmd.PersistentFlags().String(configkey.MQTTPwd, "", "")
	cmd.PersistentFlags().Bool(configkey.MQTTInsecureSkipVerify, false, "mqtt ssl 连接跳过 TLS 证书校验（自签证书场景显式开启）")
	// netkit
	cmd.PersistentFlags().String(configkey.NetPort, "", "")

	// softether
	cmd.PersistentFlags().String(configkey.SoftEtherHost, "", "")
	cmd.PersistentFlags().String(configkey.SoftEtherPort, "", "")
	cmd.PersistentFlags().String(configkey.SoftEtherPwd, "", "")
	cmd.PersistentFlags().String(configkey.SoftEtherOpenVpnPort, "", "")

	// rustfs/aws s3
	cmd.PersistentFlags().String(configkey.FSEndpoint, "127.0.0.1:9000", "")
	cmd.PersistentFlags().String(configkey.FSAccessKey, "", "")
	cmd.PersistentFlags().String(configkey.FSSecret, "", "")
}

func bind(cmd *cobra.Command) {
	// 所有类型flag
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
}
