package main

import (
	"encoding/json"
	"log"
	"net/smtp"
	"os"
	"os/user"
)

type accountConfig struct {
	URI        string
	NameFormat *string `json:",omitempty"`
	UserAgent  *string `json:",omitempty"`
}

type config struct {
	Path        string `json:"-"`
	Recipient   *string
	FromAddress string
	NameFormat  string
	UserAgent   string
	SmtpAuth    smtp.Auth
	SmtpServer  *string
	LogLevel    *string
	Accounts    map[string]accountConfig
}

func (conf *config) String() string {
	b, err := json.Marshal(conf)
	if err != nil {
		log.Panicln("Can't Marshal config")
	}
	return string(b)
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
	var conf = &config{FromAddress: from, NameFormat: "{feed.name}: {feed-title}", UserAgent: "grue/{version}"}
	return conf, nil
}

func writeDefConfig(path string) (*config, error) {
	conf, err := makeDefConfig()
	if err != nil {
		return nil, err
	}
	conf.Path = path
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
	conf.Path = path
	return conf, nil
}

func (conf *config) addAccount(name, uri string) error {
	if conf.Accounts == nil {
		conf.Accounts = make(map[string]accountConfig)
	}
	conf.Accounts[name] = accountConfig{URI: uri}
	// TODO: Use ioutil.TempFile and os.Rename to make this atomic
	return conf.write(conf.Path)
}
