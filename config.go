package main

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"os/user"
)

type accountConfig struct {
	name string
	url  string
}

type config struct {
	Recipient  *string
	From       string
	NameFormat string
	UserAgent  string
	SmtpAuth   smtp.Auth
	SmtpServer *string
	LogLevel   *string
	Accounts   map[string]accountConfig
}

func (conf *config) String() (str string) {
	var recipient, smtpserver, loglevel = "<nil>", "<nil>", "<nil>"
	str = "&{Recipient: "
	if conf.Recipient != nil {
		recipient = *conf.Recipient
	}
	if conf.SmtpServer != nil {
		smtpserver = *conf.SmtpServer
	}
	if conf.LogLevel != nil {
		loglevel = *conf.LogLevel
	}
	return fmt.Sprintf("&{Recipient: %v, From: %v, NameFormat: %v, SmtpAuth: %v, SmtpServer: %v, LogLevel: %v, Accounts: %v}", recipient, conf.From, conf.NameFormat, conf.SmtpAuth, smtpserver, loglevel, conf.Accounts)
}

func defaultFrom() (string, error) {
	cur, err := user.Current()
	if err != nil {
		return "", err
	}
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	from := cur.Username + "@" + host
	return from, nil
}

func (conf *config) write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(conf)
	return err
}

func makeDefConfig() (*config, error) {
	from, err := defaultFrom()
	if err != nil {
		return nil, err
	}
	var conf = &config{From: from, NameFormat: "{name}: {title}"}
	return conf, nil
}

func writeDefConfig(path string) (*config, error) {
	conf, err := makeDefConfig()
	if err != nil {
		return nil, err
	}
	return conf, conf.write(path)
}

func readConfig(path string) (*config, error) {
	var conf *config = new(config)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return writeDefConfig(path)
	} else if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
