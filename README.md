# notico

notico is Slack event notification bot.

## Usage

1. Create a bot account.
2. Invite the account to a channel.
3. Run `notico` with `SLACK_TOKEN` (owned by bot account) and channel (default `#admins`).

```
$ export SLACK_TOKEN=xxxxxx
$ notico [-channel "#foo"]
```

## Supported Events

- channel_created
- channel_deleted
- channel_rename
- channel_archive
- channel_unarchive
- channel_left
- team_join
- bot_added

## Options

```
Usage of notico
  -auto-archive
    	Archive the channel which includes nobody automatically
  -channel string
    	Channel to post notification message (default "#admins")
  -version
    	Show versrion
```

## LICENSE

The MIT License (MIT)

Copyright (c) 2016 KAYAC Inc.
