package zhuli

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TwilioConfig struct {
	AccountSID string
	AuthToken  string
}

type Twilio struct {
	AccountSID string
	AuthToken  string
}

func NewTwilio(twilioConfig *TwilioConfig) *Twilio {
	return &Twilio{twilioConfig.AccountSID, twilioConfig.AuthToken}
}

func (twilio *Twilio) SendText(sourceNumber string, destinationNumber string, message string) {
	postURL := "https://api.twilio.com/2010-04-01/Accounts/" + twilio.AccountSID + "/Messages.json"
	client := &http.Client{}

	data := url.Values{}
	data.Set("From", sourceNumber)
	data.Set("Body", message)
	data.Set("To", destinationNumber)

	req, _ := http.NewRequest("POST", postURL, strings.NewReader(data.Encode()))
	req.SetBasicAuth(twilio.AccountSID, twilio.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		fmt.Println(resp.Status)
	}
}
