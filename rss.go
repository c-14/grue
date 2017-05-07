package main

import (
	"fmt"
	"github.com/c-14/grue/config"
	"github.com/mmcdole/gofeed"
	"gopkg.in/gomail.v2"
	"math"
	"time"
)

type RSSFeed struct {
	config      config.AccountConfig
	LastFetched int64               `json:",omitempty"`
	LastQueried int64               `json:",omitempty"`
	NextQuery   int64               `json:",omitempty"`
	Tries       int                 `json:",omitempty"`
	GUIDList    map[string]struct{} `json:",omitempty"`
}

func hasNewerDate(item *gofeed.Item, lastFetched int64) (bool, time.Time) {
	if item.PublishedParsed != nil && item.PublishedParsed.Unix() > lastFetched {
		return true, *item.PublishedParsed
	} else if item.UpdatedParsed != nil && item.UpdatedParsed.Unix() > lastFetched {
		return true, *item.UpdatedParsed
	} else if date, exists := item.Extensions["dc"]["date"]; exists {
		dateParsed, err := time.Parse(time.RFC3339, date[0].Value)
		if err != nil {
			fmt.Printf("Can't parse (%v) as dc:date for (%v)\n", date, item.Link)
			return false, time.Now()
		}
		if dateParsed.Unix() > lastFetched {
			return true, dateParsed
		}
	}
	return false, time.Now()
}

func fetchFeed(fp *gofeed.Parser, ch chan *gomail.Message, feedName string, account *RSSFeed, config *config.GrueConfig) {
	// if account.UserAgent != nil {
	// 	feed.SetUserAgent(*account.UserAgent)
	// }
	now := time.Now()
	if account.NextQuery > now.Unix() {
		return
	}
	feed, err := fp.ParseURL(account.config.URI)
	account.LastQueried = now.Unix()
	if err != nil {
		if account.Tries > 0 {
			account.NextQuery = now.Add(time.Duration(math.Exp2(float64(account.Tries+4))) * time.Minute).Unix()
		}
		account.Tries++
		fmt.Printf("Caught error when parsing %s: %s\n", account.config.URI, err)
		return
	}
	account.NextQuery = 0
	account.Tries = 0
	guids := account.GUIDList
	if float64(len(guids)) > 1.2*float64(len(feed.Items)) {
		account.GUIDList = make(map[string]struct{})
	}
	for _, item := range feed.Items {
		// fmt.Printf("Item: %v (%v)\n\t %v: %v %v\n", item.Title, item.Link, item.GUID, item.Updated, item.Published)
		// b, _ := json.MarshalIndent(item, "", "  ")
		// fmt.Printf("Item: %v\n", string(b))
		if newer, date := hasNewerDate(item, account.LastFetched); newer {
			// fmt.Printf("New item [%v]: %v (%v)\n", date, item.Title, item.Link)
			e := createEmail(feedName, feed.Title, item, date, account.config, config)
			e.Send(ch)
		} else {
			_, exists := guids[item.GUID]
			if !exists {
				// fmt.Printf("New uid: %v (%v)\n", item.Title, item.Link)
				e := createEmail(feedName, feed.Title, item, date, account.config, config)
				e.Send(ch)
			}
			account.GUIDList[item.GUID] = struct{}{}
		}
	}
	account.LastFetched = account.LastQueried
	return
}

func fetchFeeds(ret chan error, conf *config.GrueConfig, init bool) {
	hist, err := ReadHistory()
	if err != nil {
		ret <- err
		close(ret)
		return
	}
	ch := make(chan *gomail.Message, 5)
	if !init {
		go startDialing(ch, ret, conf)
	} else {
		go func() {
			for range ch {
			}
		}()
	}

	fp := gofeed.NewParser()
	for name, accountConfig := range conf.Accounts {
		account, exist := hist.Feeds[name]
		account.config = accountConfig
		if !exist {
			account.GUIDList = make(map[string]struct{})
		}
		fetchFeed(fp, ch, name, &account, conf)
		hist.Feeds[name] = account
	}
	close(ch)
	ret <- hist.Write()
	close(ret)
}
