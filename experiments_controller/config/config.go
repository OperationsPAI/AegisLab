package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init 初始化配置
func Init(configPath string) {
	viper.SetConfigName("config")   // 配置文件名称（不含后缀）
	viper.SetConfigType("toml")     // 配置文件类型
	viper.AddConfigPath(configPath) // 配置文件路径
	viper.AddConfigPath(".")        // 当前目录

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("读取配置文件失败: %v", err)
	}
	logrus.Println("配置文件加载成功")
}

// Get 获取配置项的值
func Get(key string) interface{} {
	return viper.Get(key)
}

// GetString 获取字符串类型的配置项
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt 获取整数类型的配置项
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool 获取布尔类型的配置项
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetFloat64 获取浮点数类型的配置项
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetStringSlice 获取字符串列表类型的配置项
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetIntSlice 获取整数列表类型的配置项
func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

// GetMap 获取映射类型的配置项
func GetMap(key string) map[string]interface{} {
	return viper.GetStringMap(key)
}

// GetList 获取任意列表类型的配置项
func GetList(key string) []interface{} {
	value := viper.Get(key)
	if value == nil {
		return nil
	}
	if list, ok := value.([]interface{}); ok {
		return list
	}
	return nil
}
