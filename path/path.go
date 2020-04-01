package path

// Env 路径环境配置
type Env struct {
	localDir    string
	serviceName string
}

// PathMap 媒体路径映射表
var PathMap map[string]Env
