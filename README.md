# Chat bot for Ninjam server and Telegram

Chatbot for cross-chat messaging (between two or more Ninjam servers, Ninjam server and Telegram chat) and informating about Ninjam servers status in Telegram channel.

## Configuration

Config file support 1 or more Ninjam servers (see config.example.yaml) and one Telegram bot account.
You must get token for Telegram bot and, for full cross-chat support, bot must have full access to messages in channel (it must be admin and no-private bot mode).
Chat ID you can get from app log after adding bot to channel.

## Build

Required Go 1.8+

Linux:

```
make
```

Windows:

```
go get
go build
```

## Start

```
ninjam-chatbot -c config.yaml
```

