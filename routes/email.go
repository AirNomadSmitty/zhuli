package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/airnomadsmitty/zhuli/utils"
	"google.golang.org/api/gmail/v1"
)

type GmailSubRequestPayload struct {
	Message struct {
		Data      string `json:"data"`
		MessageID int    `json:"message_id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}
type EmailController struct {
	Srv           *gmail.Service
	LastHistoryID uint64
	AccountSID    string
	AuthToken     string
	PhoneNumber   string
}

func NewEmailController(srv *gmail.Service, lastHistoryID uint64, accountSID string, authToken string, phoneNumber string) *EmailController {
	return &EmailController{srv, lastHistoryID, accountSID, authToken, phoneNumber}
}

func (cont *EmailController) Post(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var response GmailSubRequestPayload
	err := decoder.Decode(&response)
	if err != nil {
		panic(err)
	}

	cont.processEmail(response)
	res.Write(nil)
}

func (cont *EmailController) Get(res http.ResponseWriter, req *http.Request) {
	cont.processEmail(GmailSubRequestPayload{})
	res.Write([]byte("OK"))
}

func (cont *EmailController) processEmail(pushedData GmailSubRequestPayload) {
	historyList, err := cont.Srv.Users.History.List("me").StartHistoryId(cont.LastHistoryID).HistoryTypes("messageAdded").Do()
	if err != nil {
		utils.See(err)
	}
	for _, history := range historyList.History {
		for _, historyMessage := range history.Messages {
			message, err := cont.Srv.Users.Messages.Get("me", historyMessage.Id).Do()
			if err != nil {
				utils.See(err)
			}
			if isSignupEmail(message) {
				cont.handleBookingMessage(message)
			}
		}
	}
}

func isSignupEmail(message *gmail.Message) bool {
	headers := message.Payload.Headers
	for _, header := range headers {
		if header.Name == "Subject" && strings.Contains(header.Value, "Booked") {
			return true
		}
	}

	return false
}

func parseMessagePart(message *gmail.MessagePart) ([]byte, error) {
	var body []byte
	var err error
	if message.Body != nil && message.Body.Data != "" {
		body, err = base64.URLEncoding.DecodeString(message.Body.Data)
	} else {
		var partBody []byte
		for _, part := range message.Parts {
			partBody, err = parseMessagePart(part)
			if err != nil {
				return body, err
			}
			body = append(body, partBody...)
		}
	}

	return body, err
}

func (cont *EmailController) handleBookingMessage(message *gmail.Message) {
	messageData, _ := parseMessagePart(message.Payload)

	regex, _ := regexp.Compile(`spot ([\d]{1,2}) with [\w ]+ on (Mon Mar 2nd 5:45 PM)`)
	matches := regex.FindStringSubmatch(string(messageData))

	postUrl := "https://api.twilio.com/2010-04-01/Accounts/" + cont.AccountSID + "/Messages.json"
	text := "David has signed up for spot " + matches[1] + " in the class on " + matches[2]
	utils.See(matches)

	data := url.Values{}
	data.Set("To", "")
	data.Set("From", cont.PhoneNumber)
	data.Set("Body", text)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", postUrl, strings.NewReader(data.Encode()))
	req.SetBasicAuth(cont.AccountSID, cont.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		fmt.Println(resp.Status)
	}

}
