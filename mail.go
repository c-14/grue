package main

import (
	"fmt"
	"github.com/c-14/grue/config"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"gopkg.in/gomail.v2"
	"strconv"
	"strings"
	"time"
)

type Email struct {
	FromName    string
	FromAddress string
	Recipient   string
	Date        time.Time
	Subject     string
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

func (email *Email) Send(ch chan *gomail.Message) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", email.FromAddress, email.FromName)
	m.SetHeader("To", email.Recipient)
	m.SetHeader("Subject", email.Subject)
	m.SetDateHeader("Date", email.Date)
	m.SetDateHeader("X-Date", time.Now())
	m.SetHeader("X-RSS-URI", email.ItemURI)
	bodyPlain, err := html2text.FromString(email.Body)
	if err != nil {
		fmt.Printf("Failed to parse text as HTML: %v", email.Subject)
		m.SetBody("text/html", email.Body)
	} else {
		m.SetBody("text/plain", bodyPlain)
	}
	ch <- m
}

func createEmail(feedName string, feedTitle string, item *gofeed.Item, date time.Time, account config.AccountConfig, conf *config.GrueConfig) *Email {
	email := new(Email)
	email.setFrom(feedName, feedTitle, account, conf)
	email.Recipient = conf.Recipient
	email.Subject = item.Title
	email.Date = date
	email.ItemURI = item.Link
	email.Body = item.Description
	return email
}

func startDialing(messages chan *gomail.Message, ret chan error, conf *config.GrueConfig) {
	var d *gomail.Dialer
	var hostname string
	var port int
	var err error
	if conf.SmtpServer != nil {
		parts := strings.Split(*conf.SmtpServer, ":")
		if len(parts) > 2 {
			ret <- fmt.Errorf("%s not a valid hostname\n", *conf.SmtpServer)
			return
		} else if len(parts) == 1 {
			hostname = parts[0]
			port = 587
		} else {
			hostname = parts[0]
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				ret <- fmt.Errorf("Failed to parse port: %v\n", err)
				return
			}
		}
	} else {
		hostname = "localhost"
		port = 25
	}
	if conf.SmtpUser != nil {
		d = gomail.NewDialer(hostname, port, *conf.SmtpUser, *conf.SmtpPass)
	} else {
		d = gomail.NewDialer(hostname, port, "", "")
	}

	s, err := d.Dial()
	if err != nil {
		ret <- err
		return
	}
	for m := range messages {
		err = gomail.Send(s, m)
		if err != nil {
			ret <- err
		}
	}

	ret <- s.Close()
}
