package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/c-14/grue/config"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"gopkg.in/gomail.v2"
)

type Email struct {
	FromName    string
	SenderName  string
	FromAddress string
	Recipient   string
	Date        time.Time
	Subject     string
	UserAgent   string
	ListId      string
	FeedURL     string
	ItemURI     string
	Body        string
}

func (email *Email) setFrom(feedName string, feed *gofeed.Feed, item *gofeed.Item, account config.AccountConfig, conf *config.GrueConfig) {
	var author gofeed.Person
	if item.Author != nil {
		author = *item.Author
	} else if feed.Author != nil {
		author = *feed.Author
	}
	if author.Name == "" {
		author.Name = feedName
	}
	r := strings.NewReplacer("{name}", feedName, "{title}", feed.Title,
		"{author}", author.Name)
	if account.NameFormat != nil {
		email.FromName = r.Replace(*account.NameFormat)
	} else {
		email.FromName = r.Replace(conf.NameFormat)
	}
	if author.Email == "" {
		email.FromAddress = conf.FromAddress
	} else {
		email.FromAddress = author.Email
		email.SenderName = conf.FromAddress
	}
}

func (email *Email) setUserAgent(conf *config.GrueConfig) {
	if conf.UserAgent != "" {
		r := strings.NewReplacer("{version}", version)
		email.UserAgent = r.Replace(conf.UserAgent)
	}
}

// hash is a utility function to create a string representation of the
// hash of the input string.
func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%d", h.Sum32())
}

// See RFC2919 for details on List-Id header
func (email *Email) setListId(feedName, feedURI string, conf *config.GrueConfig) {
	if conf.ListIdFormat == "" {
		return
	}
	var host string
	if u, err := url.Parse(email.FeedURL); err == nil {
		host = u.Hostname()
	}
	r := strings.NewReplacer("{name}", feedName, "{urihash}", hash(feedURI),
		"{namehash}", hash(feedName), "{host}", host)
	email.ListId = r.Replace(conf.ListIdFormat)
}

func (email *Email) Send() error {
	m := email.format()
	return gomail.Send(gomail.SendFunc(sendMail), m)
}

func (email *Email) format() *gomail.Message {
	var err error

	m := gomail.NewMessage()
	m.SetAddressHeader("From", email.FromAddress, email.FromName)
	if email.SenderName != "" {
		// if UserAgent is "", name portion of address is omitted
		m.SetAddressHeader("Sender", email.FromAddress, email.UserAgent)
	}
	m.SetHeader("To", email.Recipient)
	m.SetHeader("Subject", email.Subject)
	m.SetDateHeader("Date", email.Date)
	m.SetDateHeader("X-Date", time.Now())
	if email.UserAgent != "" {
		m.SetHeader("User-Agent", email.UserAgent)
	}
	if email.ListId != "" {
		m.SetHeader("List-Id", email.ListId)
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

func createEmail(feedName string, feed *gofeed.Feed, item *gofeed.Item, date time.Time, account config.AccountConfig, conf *config.GrueConfig) *Email {
	email := new(Email)
	email.setFrom(feedName, feed, item, account, conf)
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
	email.setListId(feedName, account.URI, conf)
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
