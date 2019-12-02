# grue

Go RSS Uranium Email

A simple RSS client that parses feeds and then sends them as emails via SMTP.

Current Status:

	Release - 0.2.2
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

* Remove a Feed from Config:
```
grue delete <name>
```

* List Feeds in Config:
```
grue list
```

* Fetch Feeds as cron job:
```
*/5 * * * *		grue fetch
```
