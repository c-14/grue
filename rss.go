package main

import (
	"c-14/grue/config"
	"fmt"
	"github.com/mmcdole/gofeed"
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

func fetchFeed(fp *gofeed.Parser, account *RSSFeed) {
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
	// TODO: make configurable
	if len(guids) > 100 {
		account.GUIDList = make(map[string]struct{})
	}
	for _, item := range feed.Items {
		if item.UpdatedParsed != nil || item.PublishedParsed != nil {
			if item.PublishedParsed != nil && item.PublishedParsed.Unix() > account.LastFetched {
				// fmt.Printf("New item: %v (%v)\n", item.Title, item.Link)
			} else if item.UpdatedParsed != nil && item.UpdatedParsed.Unix() > account.LastFetched {
				// fmt.Printf("Updated item: %v (%v)\n", item.Title, item.Link)
			}
		} else {
			_, exists := guids[item.GUID]
			if !exists {
				// fmt.Printf("New item: %v (%v)\n", item.Title, item.Link)
			}
			account.GUIDList[item.GUID] = struct{}{}
		}
	}
	account.LastFetched = account.LastQueried
	return
}

func fetchFeeds(conf *config.GrueConfig) error {
	hist, err := ReadHistory()
	if err != nil {
		return err
	}
	fp := gofeed.NewParser()
	for name, accountConfig := range conf.Accounts {
		account, exist := hist.Feeds[name]
		account.config = accountConfig
		if !exist {
			account.GUIDList = make(map[string]struct{})
		}
		fetchFeed(fp, &account)
		hist.Feeds[name] = account
	}
	return hist.Write()
}
