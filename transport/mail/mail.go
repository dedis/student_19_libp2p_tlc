package mail

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	netmail "net/mail"
	"time"
)

const mailSendServer = "mail.localhost.localdomain"
const mailReceiveServer = "localhost.localdomain"

type mail struct {
	username    string
	password    string
	recentIndex uint32
	addressBook []string
	imapSession *client.Client
}

// Broadcast sends mail to all addresses inside AddressBook
func (m *mail) Broadcast(msgBytes []byte) {
	fmt.Println(m.username, msgBytes)
	SendMail(m.username, m.addressBook, "", msgBytes, m.password)
}

// Send sends a message to a node with specific id
func (m *mail) Send(msgBytes []byte, id int) {
	// Using Broadcast
	m.Broadcast(msgBytes)
}

// Receive gets new mail from inbox
func (m *mail) Receive() *[]byte {
	msgBytes := GetMailSubject(m.imapSession, m.recentIndex)
	if msgBytes == nil {
		time.Sleep(2 * time.Second)
		return nil
	}
	fmt.Println("received :", m.username, msgBytes)
	m.recentIndex += 1
	return &msgBytes
}
func (c *mail) Disconnect() {
}

func (c *mail) Reconnect(topic string) {
}

// SendMail sends a mail from a user to several users
func SendMail(from string, to []string, subject string, body []byte, password string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", string(body))
	m.SetBody("text/plain", string(body))

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
	if mbox.Messages == 0 || index > mbox.Messages {
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

	fmt.Printf("message in mailbox %s with index %d:", username, index)
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

	m, err := netmail.ReadMessage(r)
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
	fmt.Println("Body:", string(body))
	return body
}

func GetMailSubject(c *client.Client, index uint32) []byte {
	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	if mbox.Messages == 0 || index > mbox.Messages {
		fmt.Printf("Error : %v\n", "No message with that index")
		return nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(index, index)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	if err := <-done; err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	msg := <-messages
	fmt.Println([]byte(msg.Envelope.Subject))
	return []byte(msg.Envelope.Subject)
}

func StartImapSession(username string, password string) *client.Client {
	// Connect to server
	c, err := client.DialTLS("localhost.localdomain:993", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	// TODO Don't logout here and login again. Do it only once
	// Don't forget to logout
	// defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	return c
}
