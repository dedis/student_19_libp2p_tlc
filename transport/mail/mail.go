package mail

import (
	"github.com/dedis/student_19_libp2p_tlc/model"

	"crypto/tls"
	"gopkg.in/gomail.v2"
)

const mailServer = "mail.localhost.localdomain"

type mail struct {
	username    string
	password    string
	addressBook []string
}

// Broadcast Uses PubSub publish to broadcast messages to other peers
func (m *mail) Broadcast(msg model.Message) {

}

func (m *mail) Send(msg model.Message, id int) {

}

// Receive gets message from PubSub in a blocking way
func (c *mail) Receive() *model.Message {
	return nil
}

func SendMail(from string, to []string, subject string, body []byte, password string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", string(body))

	d := gomail.NewDialer(mailServer, 25, from, password)
	// We are using self-signed certificates so we must skip verification
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
