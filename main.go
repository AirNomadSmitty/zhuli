package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/airnomadsmitty/zhuli/zhuli"
)

func main() {
	configFile, err := os.Open("config.json")
	defer configFile.Close()
	if err != nil {
		panic(err)
	}

	var config zhuli.ZhuliConfig
	configBytes, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(configBytes, &config)

	zhuli := zhuli.NewZhuli(&config)
	zhuli.DoTheThing()
}
