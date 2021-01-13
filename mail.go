package main

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/c-14/grue/config"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"gopkg.in/gomail.v2"
)

type Email struct {
	FromName    string
	FromAddress string
	Recipient   string
	Date        time.Time
	Subject     string
	UserAgent   string
	FeedURL     string
	ItemURI     string
	Body        string
}

func (email *Email) setFrom(feedName string, feedTitle string, account config.AccountConfig, conf *config.GrueConfig) {
	r := strings.NewReplacer("{name}", feedName, "{title}", feedTitle)
	if account.NameFormat != nil {
		email.FromName = r.Replace(*account.NameFormat)
	} else {
		email.FromName = r.Replace(conf.NameFormat)
	}
	email.FromAddress = conf.FromAddress
}

func (email *Email) setUserAgent(conf *config.GrueConfig) {
	if conf.UserAgent != "" {
		r := strings.NewReplacer("{version}", version)
		email.UserAgent = r.Replace(conf.UserAgent)
	}
}

func (email *Email) Send(sender gomail.Sender) error {
	m := email.format()
	return gomail.Send(sender, m)
}

func (email *Email) format() *gomail.Message {
	var err error

	m := gomail.NewMessage()
	m.SetAddressHeader("From", email.FromAddress, email.FromName)
	m.SetHeader("To", email.Recipient)
	m.SetHeader("Subject", email.Subject)
	m.SetDateHeader("Date", email.Date)
	m.SetDateHeader("X-Date", time.Now())
	if email.UserAgent != "" {
		m.SetHeader("User-Agent", email.UserAgent)
	}
	m.SetHeader("X-RSS-Feed", email.FeedURL)
	m.SetHeader("X-RSS-URI", email.ItemURI)
	bodyPlain, err := html2text.FromString(email.Body)
	if err != nil {
		fmt.Printf("Failed to parse text as HTML: %v", email.Subject)
		m.SetBody("text/html", email.Body)
	} else {
		m.SetBody("text/plain", bodyPlain)
	}
	return m
}

func createEmail(feedName string, feedTitle string, item *gofeed.Item, date time.Time, account config.AccountConfig, conf *config.GrueConfig) *Email {
	email := new(Email)
	email.setFrom(feedName, feedTitle, account, conf)
	email.Recipient = conf.Recipient
	email.Subject = item.Title
	email.Date = date
	email.setUserAgent(conf)
	email.FeedURL = account.URI
	email.ItemURI = item.Link
	if item.Content != "" {
		email.Body = item.Content
	} else {
		email.Body = item.Description
	}
	return email
}

type SendmailSender struct{}

func (sender SendmailSender) Send(from string, to []string, msg io.WriterTo) error {
	cmd := exec.Command("sendmail", "-oi", "-t")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	_, err = msg.WriteTo(stdin)
	if err != nil {
		stdin.Close()
		return err
	}
	stdin.Close()
	return cmd.Wait()
}

type SmtpSender struct {
	dialer *gomail.Dialer
}

func setupDialer(server string, user, pass *string) (gomail.Sender, error) {
	var sender SmtpSender
	var hostname string
	var port int
	var err error

	parts := strings.Split(server, ":")
	if len(parts) > 2 {
		return nil, fmt.Errorf("%s not a valid hostname\n", server)
	} else if len(parts) == 1 {
		hostname = parts[0]
		port = 587
	} else {
		hostname = parts[0]
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse port: %v\n", err)
		}
	}

	if user != nil {
		sender.dialer = gomail.NewDialer(hostname, port, *user, *pass)
	} else {
		sender.dialer = gomail.NewDialer(hostname, port, "", "")
	}

	return sender, nil
}

func (sender SmtpSender) Send(from string, to []string, msg io.WriterTo) error {
	s, err := sender.dialer.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Send(from, to, msg)
}

func setupMailer(conf *config.GrueConfig) (gomail.Sender, error) {
	if conf.SmtpServer != nil {
		return setupDialer(*conf.SmtpServer, conf.SmtpUser, conf.SmtpPass)
	}
	return SendmailSender{}, nil
}
