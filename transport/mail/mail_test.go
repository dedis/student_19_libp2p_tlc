package mail

import "testing"

func TestSendMail(t *testing.T) {
	usernames := []string{"a@localhost.localdomain", "b@localhost.localdomain", "c@localhost.localdomain", "d@localhost.localdomain", "e@localhost.localdomain"}
	SendMail(usernames[0], usernames, "test send mail function", []byte("Testing from mail package"), "apassword")
}
