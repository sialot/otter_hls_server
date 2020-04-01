package config

import (
	"flag"
	"fmt"

	"github.com/kylelemons/go-gypsy/yaml"
)

// SysConfig 配置文件对象
var SysConfig *yaml.File

//配置文件读取
func loadConfig(configFilePath string) {
	var err error
	SysConfig, err = yaml.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Config file %s not exist! \n", configFilePath)
	}
}

// InitConfig 准备配置文件
func InitConfig() {

	// 配置文件地址
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config/config.yaml", "config file path")
	flag.Parse()

	// 加载配置文件
	loadConfig(configFilePath)
}

//
