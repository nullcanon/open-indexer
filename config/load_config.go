package config

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
	"strings"
)
func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
	//刚才设置的环境变量 想要生效 我们必须得重启goland
}
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))  //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}
var Global *ServerConfig
func InitConfig(){
	//从配置文件中读取出对应的配置
	debug := GetEnvInfo("Mall_DEBUG")
	configFilePrefix := "config"
	//fmt.Println("path=",GetCurrentDirectory())
	//configFileName := fmt.Sprintf("auth/%s-pro.yaml", configFilePrefix)
	configFileName := fmt.Sprintf("./%s.yaml", configFilePrefix)
	if debug {
		configFileName = fmt.Sprintf("./%s-debug.yaml", configFilePrefix)
	}

	v := viper.New()
	//文件的路径如何设置
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	//这个对象如何在其他文件中使用 - 全局变量
	if err := v.Unmarshal(&Global); err != nil {
		panic(err)
	}
	//fmt.Printf("global +%v\n",Global)
	zap.S().Infof("配置信息: %v", Global)

	//fmt.Println(&Global)
}