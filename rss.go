package main

import (
	"fmt"
	"github.com/c-14/grue/config"
	"github.com/mmcdole/gofeed"
	"math"
	"time"
)

type FeedParser struct {
	parser   *gofeed.Parser
	messages chan *Email
	sem      chan int
	finished chan int
}

type RSSFeed struct {
	config      config.AccountConfig
	LastFetched int64               `json:",omitempty"`
	LastQueried int64               `json:",omitempty"`
	NextQuery   int64               `json:",omitempty"`
	Tries       int                 `json:",omitempty"`
	GUIDList    map[string]struct{} `json:",omitempty"`
}

type DateType int

const (
	NoDate DateType = iota
	DateNewer
	DateOlder
)

func hasNewerDate(item *gofeed.Item, lastFetched int64) (time.Time, DateType) {
	if item.PublishedParsed != nil {
		if item.PublishedParsed.Unix() > lastFetched {
			return *item.PublishedParsed, DateNewer
		} else {
			return *item.PublishedParsed, DateOlder
		}
	} else if item.UpdatedParsed != nil {
		if item.UpdatedParsed.Unix() > lastFetched {
			return *item.UpdatedParsed, DateNewer
		} else {
			return *item.UpdatedParsed, DateOlder
		}
	} else if date, exists := item.Extensions["dc"]["date"]; exists {
		dateParsed, err := time.Parse(time.RFC3339, date[0].Value)
		if err != nil {
			fmt.Printf("Can't parse (%v) as dc:date for (%v)\n", date, item.Link)
			return time.Now(), NoDate
		}
		if dateParsed.Unix() > lastFetched {
			return dateParsed, DateNewer
		} else {
			return dateParsed, DateOlder
		}
	}
	return time.Now(), NoDate
}

func fetchFeed(fp FeedParser, feedName string, account *RSSFeed, config *config.GrueConfig) {
	// if account.UserAgent != nil {
	// 	feed.SetUserAgent(*account.UserAgent)
	// }
	now := time.Now()
	if account.NextQuery > now.Unix() {
		<-fp.sem
		fp.finished <- 1
		return
	}
	feed, err := fp.parser.ParseURL(account.config.URI)
	account.LastQueried = now.Unix()
	if err != nil {
		if account.Tries > 0 {
			account.NextQuery = now.Add(time.Duration(math.Exp2(float64(account.Tries+4))) * time.Minute).Unix()
		}
		account.Tries++
		if account.Tries > 1 {
			fmt.Printf("Caught error (#%d) when parsing %s: %s\n", account.Tries, account.config.URI, err)
		}
		<-fp.sem
		fp.finished <- 1
		return
	}
	account.NextQuery = 0
	account.Tries = 0
	guids := account.GUIDList
	if float64(len(guids)) > 1.2*float64(len(feed.Items)) {
		account.GUIDList = make(map[string]struct{})
	}
	for _, item := range feed.Items {
		_, exists := guids[item.GUID]
		date, newer := hasNewerDate(item, account.LastFetched)
		if !exists || (item.GUID == "" && newer == DateNewer) {
			e := createEmail(feedName, feed.Title, item, date, account.config, config)
			err = e.Send(fp.messages)
		}
		if err == nil {
			account.GUIDList[item.GUID] = struct{}{}
		} else {
			break
		}
	}
	if err == nil {
		account.LastFetched = time.Now().Unix()
	}

	<-fp.sem
	fp.finished <- 1
}

func fetchFeeds(ret chan error, conf *config.GrueConfig, init bool) {
	var ch chan *Email = make(chan *Email)
	var dial chan int = make(chan int)
	defer close(ret)
	hist, err := ReadHistory()
	if err != nil {
		ret <- err
		return
	}
	if !init {
		s, err := setupDialer(conf)
		if err != nil {
			ret <- err
			return
		}
		go startDialing(s, ch, dial, ret)
	} else {
		go func() {
			for m := range ch {
				m.ret <- nil
			}
		}()
		close(dial)
	}

	fp := FeedParser{parser: gofeed.NewParser(), messages: ch, sem: make(chan int, 10), finished: make(chan int)}
	go func() {
		for name, accountConfig := range conf.Accounts {
			fp.sem <- 1
			account, exist := hist.Feeds[name]
			if !exist {
				account = new(RSSFeed)
				account.GUIDList = make(map[string]struct{})
				hist.Feeds[name] = account
			} else if len(account.GUIDList) == 0 {
				account.GUIDList = make(map[string]struct{})
			}
			account.config = accountConfig
			go fetchFeed(fp, name, account, conf)
		}
	}()
	for range conf.Accounts {
		<-fp.finished
	}
	ret <- hist.Write()
	close(ch)
	<-dial
}
