package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

var (
	user     = ""
	password = ""

	addr = "smtp.mail.ru:25"
	host = "smtp.mail.ru"
)

var (
	templ_file   = flag.String("templ", "template/template.html", "specify the template file for the letter")
	subs_file    = flag.String("subs", "subscribers.json", "specify the subscriber list file")
	subject_text = flag.String("subject", "", "specify the subject of the letter")
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

type Subscribers struct {
	Email string
	Name  string
	Date  string
}

func GetSubscribers() []Subscribers {
	data, err := os.ReadFile(*subs_file)
	if err != nil {
		log.Println(err)
	}

	var sub []Subscribers
	err = json.Unmarshal(data, &sub)
	if err != nil {
		fmt.Println(err)
	}

	return sub
}

func ReadTemplate(subscribers Subscribers, path string) string {
	template_file, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
	}

	tmp, _ := template.New("data").Parse(string(template_file))
	if err != nil {
		log.Println(err)
	}

	sample, err := os.OpenFile("sample.html", os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
	}

	err = tmp.Execute(sample, subscribers)
	if err != nil {
		log.Println(err)
	}

	sample_text, _ := os.ReadFile("sample.html")
	return string(sample_text)
}

func BuildTemplate(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}

func SMTPSend(subject string) {
	subs := GetSubscribers()
	for i := range subs {
		body := ReadTemplate(subs[i], *templ_file)

		template := Mail{
			Sender:  user,
			To:      []string{subs[i].Email},
			Subject: subject,
			Body:    body,
		}
		message := BuildTemplate(template)

		auth := smtp.PlainAuth("", user, password, host)
		err := smtp.SendMail(addr, auth, user, []string{subs[i].Email}, []byte(message))

		if err != nil {
			log.Println(err)
		} else {
			fmt.Println("The message has been sent!")
		}
	}
}

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "How to use:")
		flag.PrintDefaults()
	}
	flag.Parse()

	SMTPSend(*subject_text)
}
