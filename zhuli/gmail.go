package zhuli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type GmailConfig struct {
	Oauth struct {
		AccessToken  string
		RefreshToken string
	}
	Topic                  string
	PushKey                string
	GoogleCloudCredentials string
}

type GmailSubRequestPayload struct {
	Message struct {
		Data      string
		MessageID string `json:"message_id"`
	}
	Subscription string
}

type GmailMessageHandler interface {
	TriggerProcessing(message *gmail.Message) bool
	ProcessEmail(message *gmail.Message)
}

type GmailMessageData struct {
	EmailAddress string
	HistoryID    uint64
}

func (payload *GmailSubRequestPayload) ParseData() GmailMessageData {
	decoded, _ := base64.URLEncoding.DecodeString(payload.Message.Data)
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	var data GmailMessageData
	decoder.Decode(&data)
	return data
}

type Gmail struct {
	Srv                  *gmail.Service
	LastHistoryID        uint64
	PushKey              string
	GmailMessageHandlers []GmailMessageHandler
	ProcessedEmails      map[string]bool
}

func NewGmail(gmailConfig *GmailConfig) *Gmail {
	config, err := google.ConfigFromJSON([]byte(gmailConfig.GoogleCloudCredentials), gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret to config: %v", err)
	}
	tok := &oauth2.Token{
		AccessToken:  gmailConfig.Oauth.AccessToken,
		RefreshToken: gmailConfig.Oauth.RefreshToken,
		Expiry:       time.Now().Add(time.Minute * time.Duration(30)),
	}
	client := config.Client(context.Background(), tok)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	gmailHandler := &Gmail{Srv: srv, PushKey: gmailConfig.PushKey}
	go gmailHandler.doGmailDailyMaintenance(gmailConfig.Topic)
	return gmailHandler
}

func (cont *Gmail) Post(res http.ResponseWriter, req *http.Request) {
	if req.URL.Query().Get("key") != cont.PushKey {
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var response GmailSubRequestPayload
	err := decoder.Decode(&response)
	if err != nil {
		panic(err)
	}
	messageData := response.ParseData()
	cont.handleNewEmails()
	cont.LastHistoryID = messageData.HistoryID
	res.Write(nil)
}

func (cont *Gmail) Get(res http.ResponseWriter, req *http.Request) {
	cont.handleNewEmails()
	res.Write([]byte("OK"))
}

func (cont *Gmail) handleNewEmails() {
	historyList, err := cont.Srv.Users.History.List("me").StartHistoryId(cont.LastHistoryID).HistoryTypes("messageAdded").Do()
	if err != nil {
		panic(err)
	}
	for _, history := range historyList.History {
		for _, historyMessage := range history.Messages {
			if cont.ProcessedEmails[historyMessage.Id] {
				continue
			}
			message, err := cont.Srv.Users.Messages.Get("me", historyMessage.Id).Do()
			cont.ProcessedEmails[message.Id] = true
			if err != nil {
				panic(err)
			}
			for _, handler := range cont.GmailMessageHandlers {
				if handler.TriggerProcessing(message) {
					handler.ProcessEmail(message)
				}
			}
		}
	}
}

func ParseMessagePart(message *gmail.MessagePart) ([]byte, error) {
	var body []byte
	var err error
	if message.Body != nil && message.Body.Data != "" {
		body, err = base64.URLEncoding.DecodeString(message.Body.Data)
	} else {
		var partBody []byte
		for _, part := range message.Parts {
			partBody, err = ParseMessagePart(part)
			if err != nil {
				return body, err
			}
			body = append(body, partBody...)
		}
	}

	return body, err
}

func (cont *Gmail) doGmailDailyMaintenance(topic string) {
	watchResponse, err := cont.Srv.Users.Watch("me", &gmail.WatchRequest{TopicName: topic}).Do()
	if err != nil {
		panic(err)
	}
	cont.LastHistoryID = watchResponse.HistoryId
	cont.ProcessedEmails = map[string]bool{}
	time.Sleep(24 * time.Hour)
}
