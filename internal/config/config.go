package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/parker/ParkerCli/pkg/logger"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	AppName     string                 `mapstructure:"app_name"`
	Version     string                 `mapstructure:"version"`
	Debug       bool                   `mapstructure:"debug"`
	Environment string                 `mapstructure:"environment"`
	Server      ServerConfig           `mapstructure:"server"`
	Database    DatabaseConfig         `mapstructure:"database"`
	Log         LogConfig              `mapstructure:"log"`
	Docker      DockerConfig           `mapstructure:"docker"`
	Paths       map[string]string      `mapstructure:"paths"`
	Settings    map[string]interface{} `mapstructure:"settings"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// DockerConfig Docker配置
type DockerConfig struct {
	Registry  string `mapstructure:"registry"`
	Namespace string `mapstructure:"namespace"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	AppName:     "myapp",
	Version:     "0.1.0",
	Debug:       false,
	Environment: "development",
	Server: ServerConfig{
		Port:         8080,
		Host:         "localhost",
		ReadTimeout:  60,
		WriteTimeout: 60,
	},
	Database: DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Name:     "myapp",
		User:     "postgres",
		Password: "",
		SSLMode:  "disable",
	},
	Log: LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	},
	Docker: DockerConfig{
		Registry:  "docker.io",
		Namespace: "myapp",
	},
	Paths: map[string]string{
		"migrations": "./migrations",
		"logs":       "./logs",
		"data":       "./data",
	},
	Settings: map[string]interface{}{
		"timeout": 30,
	},
}

var (
	v           *viper.Viper
	configFile  string
	initialized bool
)

// Init 初始化配置
func Init(cfgFile string) error {
	if initialized && cfgFile == configFile {
		return nil
	}

	v = viper.New()

	// 设置默认值
	setDefaults()

	// 如果指定了配置文件
	if cfgFile != "" {
		configFile = cfgFile
		v.SetConfigFile(cfgFile)
	} else {
		// 默认查找位置
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.parkercli")
		v.AddConfigPath("/etc/parkercli")

		// Windows特定路径
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, "AppData", "Local", "ParkerCli"))
		}
	}

	// 环境变量
	v.SetEnvPrefix("PARKERCLI")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认配置
		logger.Info("配置文件未找到，使用默认配置")
	} else {
		logger.Info("使用配置文件: %s", v.ConfigFileUsed())
	}

	initialized = true
	return nil
}

// setDefaults 设置默认配置
func setDefaults() {
	v.SetDefault("app_name", DefaultConfig.AppName)
	v.SetDefault("version", DefaultConfig.Version)
	v.SetDefault("debug", DefaultConfig.Debug)
	v.SetDefault("environment", DefaultConfig.Environment)

	v.SetDefault("server.port", DefaultConfig.Server.Port)
	v.SetDefault("server.host", DefaultConfig.Server.Host)
	v.SetDefault("server.read_timeout", DefaultConfig.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", DefaultConfig.Server.WriteTimeout)

	v.SetDefault("database.driver", DefaultConfig.Database.Driver)
	v.SetDefault("database.host", DefaultConfig.Database.Host)
	v.SetDefault("database.port", DefaultConfig.Database.Port)
	v.SetDefault("database.name", DefaultConfig.Database.Name)
	v.SetDefault("database.user", DefaultConfig.Database.User)
	v.SetDefault("database.password", DefaultConfig.Database.Password)
	v.SetDefault("database.ssl_mode", DefaultConfig.Database.SSLMode)

	v.SetDefault("log.level", DefaultConfig.Log.Level)
	v.SetDefault("log.format", DefaultConfig.Log.Format)
	v.SetDefault("log.output", DefaultConfig.Log.Output)
	v.SetDefault("log.max_size", DefaultConfig.Log.MaxSize)
	v.SetDefault("log.max_backups", DefaultConfig.Log.MaxBackups)
	v.SetDefault("log.max_age", DefaultConfig.Log.MaxAge)
	v.SetDefault("log.compress", DefaultConfig.Log.Compress)

	v.SetDefault("docker.registry", DefaultConfig.Docker.Registry)
	v.SetDefault("docker.namespace", DefaultConfig.Docker.Namespace)

	for key, value := range DefaultConfig.Paths {
		v.SetDefault(fmt.Sprintf("paths.%s", key), value)
	}

	for key, value := range DefaultConfig.Settings {
		v.SetDefault(fmt.Sprintf("settings.%s", key), value)
	}
}

// Get 获取配置值
func Get(key string) interface{} {
	return v.Get(key)
}

// GetString 获取字符串配置值
func GetString(key string) string {
	return v.GetString(key)
}

// GetInt 获取整数配置值
func GetInt(key string) int {
	return v.GetInt(key)
}

// GetBool 获取布尔配置值
func GetBool(key string) bool {
	return v.GetBool(key)
}

// GetStringMapString 获取字符串映射
func GetStringMapString(key string) map[string]string {
	return v.GetStringMapString(key)
}

// GetStringMap 获取通用映射
func GetStringMap(key string) map[string]interface{} {
	return v.GetStringMap(key)
}

// GetAll 获取全部配置
func GetAll() Config {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Error("无法解析配置: %v", err)
		return DefaultConfig
	}
	return cfg
}

// Set 设置配置值
func Set(key string, value interface{}) {
	v.Set(key, value)
}

// Save 保存配置到文件
func Save() error {
	if configFile == "" {
		configFile = "config.yaml"
	}
	return v.WriteConfigAs(configFile)
}

// Reset 重置配置为默认值
func Reset() {
	v = viper.New()
	setDefaults()
}

// ConfigFileUsed 返回当前使用的配置文件路径
func ConfigFileUsed() string {
	return v.ConfigFileUsed()
}

// ConfigFileExists 检查配置文件是否存在
func ConfigFileExists() bool {
	if configFile == "" {
		return false
	}
	_, err := os.Stat(configFile)
	return err == nil
}
