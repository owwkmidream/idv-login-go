package config

import (
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
	"idv-login-go/constants"
	"idv-login-go/logger"
	"os"
	"sync"
)

var once sync.Once
var instance *Config

var log *logrus.Logger
var configPath = "./config.toml"
var defaultConf = map[string]interface{}{
	"debug":     false,
	"host":      "service.mkey.163.com",
	"hostDNS":   "https://dns.alidns.com/resolve",
	"defaultIP": "42.186.193.21",
}

type Config struct {
	*koanf.Koanf
}

func (c *Config) Save() bool {
	bytes, err := c.Marshal(toml.Parser())
	if err != nil {
		log.Errorf("转换为字节码失败：%v", err)
		return false
	}
	str := string(bytes)
	confFile, err := os.Create(configPath)
	if err != nil {
		log.Errorf("打开配置文件失败：%v", err)
		return false
	}
	defer func(confFile *os.File) {
		confFile.Close()
	}(confFile)

	// 直接写入字符串
	_, err = confFile.WriteString(str)
	if err != nil {
		log.Errorf("写入配置文件失败：%v", err)
		return false
	}

	return true
}

func GetConfig() *Config {
	once.Do(func() {
		log = logger.GetLogger()
		instance = &Config{koanf.New(".")}
		if err := instance.Load(file.Provider(configPath), toml.Parser()); err != nil {
			log.Errorf("加载配置文件失败：%v", err)
			log.Info("将使用默认配置")

			err = instance.Load(confmap.Provider(defaultConf, "."), nil)
			if err != nil {
				log.Fatalf("加载默认配置失败：%v", err)
			}
			instance.Save()
		}
		log.Info("加载配置文件成功")

		// 改变log等级
		if instance.Bool("debug") {
			log.SetLevel(logrus.DebugLevel)
			constants.DebugMode = true
		}
		log.Infof("debug模式：%v", constants.DebugMode)
	})
	return instance
}
