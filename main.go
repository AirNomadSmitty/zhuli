package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/airnomadsmitty/zhuli/zhuli"
	"google.golang.org/api/gmail/v1"
)

type MiscConfig struct {
	PhoneNumber        string
	DestinationNumbers []string
}

func main() {
	zhuliConfig := parseZhuliConfig("config.zhuli.json")
	miscConfig := parseMiscConfig("config.misc.json")

	zhuli := zhuli.NewZhuli(&zhuliConfig)
	cyclebarHandler := cyclebarSignupHandler{miscConfig.DestinationNumbers, miscConfig.PhoneNumber, zhuli.Twilio.SendText}
	zhuli.AddGmailMessageHandler(cyclebarHandler)
	zhuli.DoTheThing()
}

func parseZhuliConfig(filename string) zhuli.ZhuliConfig {
	configFile, err := os.Open("config.zhuli.json")
	defer configFile.Close()
	if err != nil {
		panic(err)
	}

	var zhuliConfig zhuli.ZhuliConfig
	configBytes, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(configBytes, &zhuliConfig)

	return zhuliConfig
}

func parseMiscConfig(filename string) MiscConfig {
	configFile, err := os.Open("config.misc.json")
	defer configFile.Close()
	if err != nil {
		panic(err)
	}

	var miscConfig MiscConfig
	configBytes, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(configBytes, &miscConfig)

	return miscConfig

}

type cyclebarSignupHandler struct {
	DestinationNumbers []string
	SourceNumber       string
	SendText           func(sourceNumber string, destinationNumber string, message string)
}

func (cyclebarSignupHandler) TriggerProcessing(message *gmail.Message) bool {
	headers := message.Payload.Headers
	for _, header := range headers {
		if header.Name == "Subject" && strings.Contains(header.Value, "Booked") {
			return true
		}
	}

	return false
}

func (handler cyclebarSignupHandler) ProcessEmail(message *gmail.Message) {
	messageData, _ := zhuli.ParseMessagePart(message.Payload)

	regex, _ := regexp.Compile(`spot ([\d]{1,2}) with [\w ]+ on ([\w :]+(?:PM|AM))`)
	matches := regex.FindStringSubmatch(string(messageData))

	text := "This is Zhu Li, David's personal assistant. He signed up for spot " + matches[1] + " in the spin class on " + matches[2] + ". Please do not reply to this message."

	for _, number := range handler.DestinationNumbers {
		handler.SendText(handler.SourceNumber, number, text)
	}

}
