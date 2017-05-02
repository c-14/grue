package config

import (
	"encoding/json"
	"log"
	"net/smtp"
	"path"
	"os"
	"os/user"
)

type AccountConfig struct {
	URI			string
	NameFormat	*string		`json:",omitempty"`
	UserAgent	*string		`json:",omitempty"`
}

type GrueConfig struct {
	path        string
	Recipient   *string
	FromAddress string
	NameFormat  string
	UserAgent   string
	SmtpAuth    smtp.Auth
	SmtpServer  *string
	LogLevel    *string
	Accounts    map[string]AccountConfig
}

func (conf *GrueConfig) String() string {
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

func (conf *GrueConfig) write(path string) error {
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

func makeDefConfig() (*GrueConfig, error) {
	from, err := defaultFrom()
	if err != nil {
		return nil, err
	}
	var conf = &GrueConfig{FromAddress: from, NameFormat: "{feed.name}: {feed-title}", UserAgent: "grue/{version}"}
	return conf, nil
}

func writeDefConfig(path string) (*GrueConfig, error) {
	conf, err := makeDefConfig()
	if err != nil {
		return nil, err
	}
	conf.path = path
	return conf, conf.write(path)
}

func getConfigPath() string {
	cfgPath := os.Getenv("XDG_CONFIG_HOME")
	if cfgPath == "" {
		home := os.Getenv("HOME")
		if home == "" {
			panic("Can't find path to data directory")
		}
		return path.Join(os.Getenv("HOME"), ".config", "grue.cfg")
	}
	return path.Join(cfgPath, "grue.cfg")
}

func ReadConfig() (*GrueConfig, error) {
	var conf *GrueConfig = new(GrueConfig)
	var path = getConfigPath()
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
	conf.path = path
	return conf, nil
}

func (conf *GrueConfig) AddAccount(name, uri string) error {
	if conf.Accounts == nil {
		conf.Accounts = make(map[string]AccountConfig)
	}
	conf.Accounts[name] = AccountConfig{URI: uri}
	// TODO: Use ioutil.TempFile and os.Rename to make this atomic
	return conf.write(conf.path)
}
