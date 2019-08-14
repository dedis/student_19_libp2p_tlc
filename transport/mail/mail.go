package mail

import (
	"crypto/tls"
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/model"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	net_mail "net/mail"
)

const mailSendServer = "mail.localhost.localdomain"
const mailReceiveServer = "localhost.localdomain"

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

// SendMail sends a mail from a user to several users
func SendMail(from string, to []string, subject string, body []byte, password string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", string(body))

	d := gomail.NewDialer(mailSendServer, 25, from, password)
	// We are using self-signed certificates so we must skip verification
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// GetMail gets a mail with specified index from inbox
func GetMail(username string, password string, index uint32) []byte {
	// Connect to server, skipping certificate verification
	c, err := client.DialTLS(mailReceiveServer+":993", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	// Logout
	defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	// Get the message at specified index
	if mbox.Messages == 0 && index > mbox.Messages {
		fmt.Printf("Error : %v\n", "No message with that index")
		return nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(index, index)

	// Get message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	fmt.Println("message:")
	msg := <-messages
	r := msg.GetBody(section)
	if r == nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	if err := <-done; err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	m, err := net_mail.ReadMessage(r)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	header := m.Header
	fmt.Println("Date:", header.Get("Date"))
	fmt.Println("From:", header.Get("From"))
	fmt.Println("To:", header.Get("To"))
	fmt.Println("Subject:", header.Get("Subject"))

	body, err := ioutil.ReadAll(m.Body)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	fmt.Println(string(body))
	return body
}
