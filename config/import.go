package config

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
)

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), prefix)
}

func getValue(s string) string {
	return strings.TrimSpace(strings.SplitN(s, "=", 2)[1])
}

func importDefault(text string) (*GrueConfig, error) {
	conf, err := makeDefConfig()
	if err != nil {
		return nil, err
	}
	for _, str := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(str, "[") && !strings.EqualFold(str, "[DEFAULT]"):
			return conf, nil
		case hasPrefix(str, "to"):
			rec := getValue(str)
			conf.Recipient = rec
		case hasPrefix(str, "from"):
			conf.FromAddress = getValue(str)
		case hasPrefix(str, "name-format"):
			r := strings.NewReplacer("{feed.name}", "{name}", "{feed-title}", "{title}")
			conf.NameFormat = r.Replace(getValue(str))
		case hasPrefix(str, "smtp-server"):
			serv := getValue(str)
			conf.SmtpServer = &serv
		default:
			continue
		}
	}
	return conf, nil
}

func splitSections(data []byte, atEOF bool) (advance int, token []byte, err error) {
	ind := bytes.IndexRune(bytes.TrimPrefix(data, []byte("[")), '[')
	if ind == 0 {
		panic("splitSections: encountered double '[' at start of Section.")
	} else if ind > 0 {
		byte_len := len([]byte("\n"))
		advance = ind + byte_len
		token = data[:ind-byte_len]
		err = nil
		return
	} else if atEOF {
		return len(data), data, io.EOF
	} else {
		return 0, nil, nil
	}
}

func ImportCfg(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: grue import <config>")
	}
	var cfgPath string = args[0]
	file, err := os.Open(cfgPath)
	if err != nil {
		return err
	}
	defer file.Close()
	cfgData := bufio.NewScanner(file)
	cfgData.Split(splitSections)
	cfgData.Scan()

	conf, err := importDefault(cfgData.Text())
	conf.path = getConfigPath()
	conf.Accounts = make(map[string]AccountConfig)

	for cfgData.Scan() {
		var name, uri string
		str := cfgData.Text()
		lines := strings.Split(str, "\n")
		if hasPrefix(lines[0], "[feed.") && hasPrefix(lines[1], "url") {
			name = strings.TrimSuffix(strings.TrimPrefix(lines[0], "[feed."), "]")
			uri = getValue(lines[1])
		} else {
			panic("Malformated feed section")
		}
		conf.Accounts[name] = AccountConfig{URI: uri}
	}
	// TODO: Use ioutil.TempFile and os.Rename to make this atomic
	return conf.write(conf.path)
}
