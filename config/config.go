package config

import (
	"gopkg.in/yaml.v3"
	"iot-master/args"
	"log"
	"os"
)

func Existing() bool {
	return existing
}

var existing = false

//Configure 配置
type Configure struct {
	Node     string   `yaml:"node" json:"node"`
	Data     string   `yaml:"data" json:"data"`
	Web      Web      `yaml:"web" json:"web"`
	Database Database `yaml:"database" json:"database"`
	History  History  `yaml:"history" json:"history"`
	Log      Log      `yaml:"log" json:"log"`
	//Serials  []string `yaml:"serials" json:"serials"`
}

//Config 全局配置
var Config = Configure{
	Node:     "root",
	Data:     "data",
	Web:      WebDefault,
	Database: DatabaseDefault,
	History:  HistoryDefault,
	Log:      LogDefault,
}

func init() {
	Config.Node, _ = os.Hostname()
}

//Load 加载
func Load() error {
	//log.Println("加载配置")
	//从参数中读取配置文件名
	filename := args.ConfigPath

	// 如果没有文件，则使用默认信息创建
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		//return Store()
		return nil
	} else {
		y, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer y.Close()

		d := yaml.NewDecoder(y)
		err = d.Decode(&Config)
		if err != nil {
			log.Fatal(err)
			return err
		}

		existing = true

		return nil
	}
}

func Store() error {
	//log.Println("保存配置")
	//从参数中读取配置文件名
	filename := args.ConfigPath

	y, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755) //os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer y.Close()

	e := yaml.NewEncoder(y)
	defer e.Close()

	err = e.Encode(&Config)
	if err != nil {
		log.Fatal(err)
		return err
	}

	existing = true

	return nil
}
