package zhuli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/airnomadsmitty/zhuli/routes"
	"github.com/airnomadsmitty/zhuli/utils"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type ZhuliConfig struct {
	Google struct {
		Gmail struct {
			Oauth struct {
				AccessToken  string
				RefreshToken string
			}
			Topic   string
			PushKey string
		}
		GoogleCloud struct {
			Credentials string
		}
	}
	WebServer struct {
		Port int
	}
	Twilio struct {
		AccountSID         string
		AuthToken          string
		PhoneNumber        string
		DestinationNumbers []string
	}
}

type Zhuli struct {
	Config *ZhuliConfig
}

func NewZhuli(config *ZhuliConfig) *Zhuli {
	return &Zhuli{config}
}

func (zhuli *Zhuli) DoTheThing() {

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(zhuli.Config.Google.GoogleCloud.Credentials), gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	tok := &oauth2.Token{
		AccessToken:  zhuli.Config.Google.Gmail.Oauth.AccessToken,
		RefreshToken: zhuli.Config.Google.Gmail.Oauth.RefreshToken,
		Expiry:       time.Now().Add(time.Minute * time.Duration(30)),
	}
	client := config.Client(context.Background(), tok)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	err = srv.Users.Stop("me").Do()
	if err != nil {
		panic(err)
	}
	topic := zhuli.Config.Google.Gmail.Topic
	watchResponse, err := srv.Users.Watch("me", &gmail.WatchRequest{TopicName: topic}).Do()
	if err != nil {
		panic(err)
	}
	utils.See(watchResponse.HistoryId)
	r := mux.NewRouter()
	emailController := routes.NewEmailController(srv, watchResponse.HistoryId, zhuli.Config.Twilio.AccountSID, zhuli.Config.Twilio.AuthToken, zhuli.Config.Twilio.PhoneNumber, zhuli.Config.Twilio.DestinationNumbers, zhuli.Config.Google.Gmail.PushKey)
	go doGmailDailyMaintenance(srv, topic, emailController)
	r.HandleFunc("/email", emailController.Post).Methods("POST")
	r.HandleFunc("/", emailController.Get).Methods("GET")
	http.Handle("/", r)
	fmt.Println("Starting server...")
	http.ListenAndServe(":"+strconv.Itoa(zhuli.Config.WebServer.Port), nil)
}

func doGmailDailyMaintenance(srv *gmail.Service, topic string, emailController *routes.EmailController) {
	time.Sleep(24 * time.Hour)
	watchResponse, err := srv.Users.Watch("me", &gmail.WatchRequest{TopicName: topic}).Do()
	if err != nil {
		panic(err)
	}
	emailController.LastHistoryID = watchResponse.HistoryId
	emailController.ProcessedEmails = map[string]bool{}
}
