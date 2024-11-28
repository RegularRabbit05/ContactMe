package api

import (
	"encoding/json"
	"fmt"
	mailer "github.com/wneessen/go-mail"
	"net/http"
	"net/mail"
	"os"
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

	type Mail struct {
		Origin string `json:"from"`
		Title  string `json:"subject"`
		Body   string `json:"body"`
	}

	var m Mail
	ip := "IP:" + r.RemoteAddr + "|HEADER:" + r.Header.Get("X-Forwarded-For")
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil || m.Origin == "" || m.Title == "" || m.Body == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	valid := func(email string) bool {
		_, err := mail.ParseAddress(email)
		return err == nil
	}

	if !valid(m.Origin) {
		http.Error(w, "Invalid sender", http.StatusBadRequest)
		return
	}

	body := fmt.Sprintf(messageTemplate, m.Origin, ip, m.Body)

	message := mailer.NewMsg()
	if err = message.From(os.Getenv("VAR_SMTP_FROM")); err != nil {
		http.Error(w, "Unable to send mail", http.StatusInternalServerError)
		return
	}
	if err = message.To(os.Getenv("VAR_SMTP_TO")); err != nil {
		http.Error(w, "Unable to send mail", http.StatusInternalServerError)
		return
	}

	message.Subject(m.Title)
	message.SetBodyString(mailer.TypeTextPlain, body)
	client, err := mailer.NewClient(os.Getenv("VAR_SMTP_SERVER"), mailer.WithTLSPortPolicy(mailer.TLSOpportunistic), mailer.WithSMTPAuth(mailer.SMTPAuthPlain),
		mailer.WithUsername(os.Getenv("VAR_SMTP_USER")), mailer.WithPassword(os.Getenv("VAR_SMTP_PASSWORD")))
	if err != nil {
		http.Error(w, "Connection failure", http.StatusBadGateway)
		return
	}

	if err = client.DialAndSend(message); err != nil {
		http.Error(w, "Unable to send mail", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
}
