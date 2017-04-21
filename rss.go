package main

import (
	"c-14/grue/config"
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
)

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)
}

func fetchFeed(account config.AccountConfig) error {
	feed := rss.New(5, true, chanHandler, itemHandler)
	if account.UserAgent != nil {
		feed.SetUserAgent(*account.UserAgent)
	}
	err := feed.Fetch(account.URI, nil)
	if err != nil {
		return err
	}
	return nil
}

func fetchFeeds(conf *config.GrueConfig) error {
	for name, account := range conf.Accounts {
		err := fetchFeed(account)
		if err != nil {
			return fmt.Errorf("%s (%s): %s\n", name, account.URI, err)
		}
	}
	return nil
}
