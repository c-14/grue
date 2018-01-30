package main

import (
	"fmt"
	"github.com/c-14/grue/config"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"gopkg.in/gomail.v2"
	"io"
	"os/exec"
	"strings"
	"time"
)

type Email struct {
	FromName    string
	FromAddress string
	Recipient   string
	Date        time.Time
	Subject     string
	UserAgent   string
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

func (email *Email) Send() error {
	m := email.format()
	return gomail.Send(gomail.SendFunc(sendMail), m)
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
	email.ItemURI = item.Link
	if item.Description == "" {
		email.Body = item.Content
	} else {
		email.Body = item.Description
	}
	return email
}

func sendMail(from string, to []string, msg io.WriterTo) error {
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
