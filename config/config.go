package config

import (
	"flag"
	"fmt"

	"github.com/kylelemons/go-gypsy/yaml"
)

var SysConfig *yaml.File

//配置文件读取
func loadConfig(configFilePath string) {
	var err error
	SysConfig, err = yaml.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Config file %s not exist!\n", configFilePath)
	}
}

// 准备配置文件
func InitConfig() {

	// 配置文件地址
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config/config.yaml", "config file path")
	flag.Parse()

	// 加载配置文件
	loadConfig(configFilePath)
}

func Get(key string) (string, error){
	return SysConfig.Get(key)
}