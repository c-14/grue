# grue

Go RSS Uranium Email

A simple RSS client that parses feeds and then sends them as emails via SMTP.

Current Status:

	Release - 0.1.0-alpha
	Bugs - Probably

## Download

	go get github.com/c-14/grue

## Usage

* Import rss2email Config:
```
grue import rss2email.cfg
```

* Add new Feed to Config:
```
grue add <name> <url>
```

* Fetch Feeds as cron job:
```
*/5 * * * *		grue fetch
```
