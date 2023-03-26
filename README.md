# News Feed Bot

Bot for Telegram that gets and posts news to a channel.

# TODO

- [x] Admin commands
- [x] Filter by keywords
- [x] Tests
- [x] golangci-lint
- [x] Docker

# Backlog

- [ ] More types of resources — not only RSS
- [x] Summary for the article
- [ ] Dynamic source priority (based on 👍 and 👎 reactions) — currently blocked by Telegram Bot API
- [ ] Article types: text, video, audio
- [ ] De-duplication — filter articles with the same title and author
- [ ] Low quality articles filter — need research
	- Ban by author? 
	- Check article length — not working with audio/video posts, but it will be fixed after article type implementation
