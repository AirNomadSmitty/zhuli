package zhuli

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/airnomadsmitty/zhuli/routes"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type Zhuli struct {
	port string
}

func NewZhuli(port string) *Zhuli {
	return &Zhuli{port}
}

func (zhuli *Zhuli) DoTheThing() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	tok := &oauth2.Token{
		AccessToken:  "accesstoken",
		RefreshToken: "refreshtoken",
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
	_, err = srv.Users.Watch("me", &gmail.WatchRequest{TopicName: "topic"}).Do()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	emailController := routes.NewEmailController(srv, 4706)
	r.HandleFunc("/email", emailController.Post).Methods("POST")
	r.HandleFunc("/", emailController.Get).Methods("GET")
	http.Handle("/", r)
	fmt.Println("Starting server...")
	http.ListenAndServe(zhuli.port, nil)

}
