package zhuli

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ZhuliConfig struct {
	Gmail     *GmailConfig
	WebServer struct {
		Port int
	}
	Twilio *TwilioConfig
}

type Zhuli struct {
	WebServerPort int
	Gmail         *Gmail
	Twilio        *Twilio
}

func NewZhuli(config *ZhuliConfig) *Zhuli {
	zhu := &Zhuli{WebServerPort: config.WebServer.Port}
	if config.Gmail != nil {
		zhu.Gmail = NewGmail(config.Gmail)
	}
	if config.Twilio != nil {
		zhu.Twilio = NewTwilio(config.Twilio)
	}
	return zhu
}

func (zhuli *Zhuli) DoTheThing() {
	r := mux.NewRouter()
	if zhuli.Gmail != nil {
		r.HandleFunc("/email", zhuli.Gmail.Post).Methods("POST")
		r.HandleFunc("/", zhuli.Gmail.Get).Methods("GET")
	}
	http.Handle("/", r)
	fmt.Println("Starting server...")
	http.ListenAndServe(":"+strconv.Itoa(zhuli.WebServerPort), nil)
}

func (zhuli *Zhuli) AddGmailMessageHandler(handler GmailMessageHandler) error {
	if zhuli.Gmail == nil {
		return errors.New("Gmail module not initialized")
	}
	zhuli.Gmail.GmailMessageHandlers = append(zhuli.Gmail.GmailMessageHandlers, handler)
	return nil
}
