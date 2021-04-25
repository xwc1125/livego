package configure

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

/*
{
  "server": [
    {
      "appname": "live",
      "live": true,
	  "hls": true,
	  "static_push": []
    }
  ]
}
*/

// Application 应用信息
type Application struct {
	AppName    string   `json:"appname" mapstructure:"appname"`         // 应用名称
	Live       bool     `json:"live" mapstructure:"live"`               //
	Hls        bool     `json:"hls" mapstructure:"hls"`                 // 是否启动hls
	Flv        bool     `json:"flv" mapstructure:"flv"`                 // 是否启动flv
	Api        bool     `json:"api" mapstructure:"api"`                 // 是否启动api
	StaticPush []string `json:"static_push" mapstructure:"static_push"` // 静态推送地址
}

type Applications []Application

type JWT struct {
	Secret    string `mapstructure:"secret"`
	Algorithm string `mapstructure:"algorithm"`
}
type ServerCfg struct {
	Level            string       `mapstructure:"level"`
	ConfigFile       string       `mapstructure:"config_file"`
	FLVArchive       bool         `mapstructure:"flv_archive"`
	ArchiveMp4       bool         `mapstructure:"archive_mp4"`       // 是否归档mp4
	ArchiveSingleton bool         `mapstructure:"archive_singleton"` // 是否归档保存单个文件
	ArchiveDir       string       `mapstructure:"archive_dir"`       // 归档目录
	FLVDir           string       `mapstructure:"flv_dir"`
	RTMPNoAuth       bool         `mapstructure:"rtmp_noauth"`
	RTMPAddr         string       `mapstructure:"rtmp_addr"`
	HTTPFLVAddr      string       `mapstructure:"httpflv_addr"`
	HLSAddr          string       `mapstructure:"hls_addr"`
	HLSKeepAfterEnd  bool         `mapstructure:"hls_keep_after_end"`
	APIAddr          string       `mapstructure:"api_addr"`
	RedisAddr        string       `mapstructure:"redis_addr"`
	RedisPwd         string       `mapstructure:"redis_pwd"`
	ReadTimeout      int          `mapstructure:"read_timeout"`
	WriteTimeout     int          `mapstructure:"write_timeout"`
	GopNum           int          `mapstructure:"gop_num"`
	JWT              JWT          `mapstructure:"jwt"`
	Server           Applications `mapstructure:"server"`
}

// default config
var defaultConf = ServerCfg{
	ConfigFile:       "livego.yaml",
	FLVArchive:       false,
	ArchiveMp4:       false,
	ArchiveSingleton: false,
	RTMPNoAuth:       false,
	RTMPAddr:         ":1935",
	HTTPFLVAddr:      ":7001",
	HLSAddr:          ":7002",
	HLSKeepAfterEnd:  false,
	APIAddr:          ":8090",
	WriteTimeout:     10,
	ReadTimeout:      10,
	GopNum:           1,
	Server: Applications{{
		AppName:    "live",
		Live:       true,
		Hls:        true,
		Flv:        true,
		Api:        true,
		StaticPush: nil,
	}},
}

var Config = viper.New()

// initLog 初始化日志
func initLog() {
	if l, err := log.ParseLevel(Config.GetString("level")); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
	}
}

func init() {
	defer Init()

	// Default config
	b, _ := json.Marshal(defaultConf)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	viper.ReadConfig(defaultConfig)
	Config.MergeConfigMap(viper.AllSettings())

	// Flags
	pflag.String("rtmp_addr", ":1935", "RTMP server listen address")
	pflag.String("httpflv_addr", ":7001", "HTTP-FLV server listen address")
	pflag.String("hls_addr", ":7002", "HLS server listen address")
	pflag.String("api_addr", ":8090", "HTTP manage interface server listen address")
	pflag.String("config_file", "livego.yaml", "configure filename")
	pflag.String("level", "info", "Log level")
	// 在流结束后维护HLS
	pflag.Bool("hls_keep_after_end", false, "Maintains the HLS after the stream ends")
	pflag.String("flv_dir", "tmp", "output flv file at flvDir/APP/KEY_TIME.flv")
	pflag.Int("read_timeout", 10, "read time out")
	pflag.Int("write_timeout", 10, "write time out")
	pflag.Int("gop_num", 1, "gop num")
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)

	// File
	Config.SetConfigFile(Config.GetString("config_file"))
	Config.AddConfigPath(".")
	err := Config.ReadInConfig()
	if err != nil {
		log.Warning(err)
		log.Info("Using default config")
	} else {
		Config.MergeInConfig()
	}

	// Environment
	replacer := strings.NewReplacer(".", "_")
	Config.SetEnvKeyReplacer(replacer)
	Config.AllowEmptyEnv(true)
	Config.AutomaticEnv()

	// Log
	initLog()

	// Print final config
	c := ServerCfg{}
	Config.Unmarshal(&c)
	log.Debugf("Current configurations: \n%# v", pretty.Formatter(c))
}

func CheckAppName(appname string) bool {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if app.AppName == appname {
			return app.Live
		}
	}
	return false
}

// GetStaticPushUrlList 获取静态推送地址
func GetStaticPushUrlList(appname string) ([]string, bool) {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if (app.AppName == appname) && app.Live {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			} else {
				return nil, false
			}
		}
	}
	return nil, false
}
