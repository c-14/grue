package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"text/tabwriter"
)

type AccountConfig struct {
	URI        string
	NameFormat *string `json:",omitempty"`
	UserAgent  *string `json:",omitempty"`
}

func (cfg AccountConfig) String() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 11, 8, 0, '\t', 0)
	fmt.Fprintf(w, "URI\t\"%s\"\n", cfg.URI)
	if cfg.NameFormat != nil {
		fmt.Fprintf(w, "Name Format\t\"%s\"\n", *cfg.NameFormat)
	}
	if cfg.UserAgent != nil {
		fmt.Fprintf(w, "User Agent\t\"%s\"\n", *cfg.UserAgent)
	}
	w.Flush()
	return b.String()
}

type GrueConfig struct {
	path        string
	Recipient   string
	FromAddress string
	NameFormat  string
	UserAgent   string
	SmtpUser    *string
	SmtpPass    *string
	SmtpServer  *string
	LogLevel    *string
	Accounts    map[string]AccountConfig
}

func (conf *GrueConfig) Lock() error {
	lock, err := os.OpenFile(conf.path+".lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	switch {
	case err == nil:
		defer lock.Close()
		_, err = lock.WriteString(string(os.Getpid()))
		return err
	case os.IsExist(err):
		return fmt.Errorf("Aborting due to existing lock on %s\n", conf.path)
	default:
		return err
	}
}

func (conf *GrueConfig) Unlock() error {
	return os.Remove(conf.path + ".lock")
}

func (conf *GrueConfig) String() string {
	b, err := json.Marshal(conf)
	if err != nil {
		panic("Can't Marshal config")
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

func (conf *GrueConfig) encode(file *os.File) error {
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(conf)
}

func (conf *GrueConfig) save() error {
	tmpfile, err := ioutil.TempFile(path.Dir(conf.path), path.Base(conf.path))
	if err != nil {
		return err
	}
	err = conf.encode(tmpfile)
	tmpfile.Close()
	if err != nil {
		os.Remove(tmpfile.Name())
		return err
	}
	return os.Rename(tmpfile.Name(), conf.path)
}
func makeDefConfig() (*GrueConfig, error) {
	from, err := defaultFrom()
	if err != nil {
		return nil, err
	}
	var conf = &GrueConfig{FromAddress: from, NameFormat: "{name}: {title}", UserAgent: "grue/{version}"}
	return conf, nil
}

func writeDefConfig(path string) (*GrueConfig, error) {
	conf, err := makeDefConfig()
	if err != nil {
		return nil, err
	}
	conf.path = path
	return conf, conf.save()
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
	conf.path = getConfigPath()
	err := conf.Lock()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(conf.path)
	if os.IsNotExist(err) {
		return writeDefConfig(conf.path)
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

func (conf *GrueConfig) AddAccount(name, uri string) error {
	if conf.Accounts == nil {
		conf.Accounts = make(map[string]AccountConfig)
	}
	conf.Accounts[name] = AccountConfig{URI: uri}
	return conf.save()
}

func (conf *GrueConfig) DeleteAccount(name string) error {
	if conf.Accounts == nil {
		return nil
	}
	if _, ok := conf.Accounts[name]; !ok {
		return fmt.Errorf("%s: account does not exist", name)
	}
	delete(conf.Accounts, name)
	return conf.save()
}
