package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

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
}

func NewEmailController(srv *gmail.Service, lastHistoryID uint64) *EmailController {
	return &EmailController{srv, lastHistoryID}
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
	historyList, err := cont.Srv.Users.History.List("me").StartHistoryId(cont.LastHistoryID).Do()
	if err != nil {
		utils.See(err)
	}
	utils.See(historyList.History)
	utils.See(historyList.HistoryId)
	for _, history := range historyList.History {
		for _, historyMessage := range history.Messages {
			message, err := cont.Srv.Users.Messages.Get("me", historyMessage.Id).Do()
			if err != nil {
				utils.See(err)
			}
			messageData, _ := parseMessagePart(message.Payload)
			fmt.Println(string(messageData))
		}
	}
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
