package api

import (
	"encoding/json"
	"fmt"
	"github.com/wneessen/go-mail"
	"log"
	"net/http"
	"os"

	"github.com/go-email-validator/go-email-validator/pkg/ev"
	"github.com/go-email-validator/go-email-validator/pkg/ev/evmail"
)

func SendMessage(w http.ResponseWriter, r *http.Request) {
	const messageTemplate = `
Message from: %s (%s)

---------------------------------

%s
`
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("VAR_CORS"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		if r := recover(); r != nil {
		}
	}()

	type Mail struct {
		Origin string `json:"from"`
		Title  string `json:"subject"`
		Body   string `json:"body"`
	}

	e := func() {
		var m Mail
		ip := "IP:" + r.RemoteAddr + "|HEADER:" + r.Header.Get("X-Forwarded-For")
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil || m.Origin == "" || m.Title == "" || m.Body == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if !ev.NewSyntaxValidator().Validate(ev.NewInput(evmail.FromString(m.Origin))).IsValid() {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := fmt.Sprintf(messageTemplate, m.Origin, ip, m.Body)

		message := mail.NewMsg()
		if err = message.From(os.Getenv("VAR_SMTP_FROM")); err != nil {
			http.Error(w, "Unable to send mail", http.StatusInternalServerError)
			log.Panicln(err)
			return
		}
		if err = message.To(os.Getenv("VAR_SMTP_TO")); err != nil {
			http.Error(w, "Unable to send mail", http.StatusInternalServerError)
			log.Panicln(err)
			return
		}

		message.Subject(m.Title)
		message.SetBodyString(mail.TypeTextPlain, body)
		client, err := mail.NewClient(os.Getenv("VAR_SMTP_SERVER"), mail.WithTLSPortPolicy(mail.TLSOpportunistic), mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(os.Getenv("VAR_SMTP_USER")), mail.WithPassword(os.Getenv("VAR_SMTP_PASSWORD")))
		if err != nil {
			http.Error(w, "Connection failure", http.StatusBadGateway)
			log.Panicln(err)
			return
		}

		if err = client.DialAndSend(message); err != nil {
			http.Error(w, "Unable to send mail", http.StatusBadGateway)
			log.Panicln(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
	e()
}
