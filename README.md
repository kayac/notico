# notico

notico is Slack event notification bot.

## Usage

1. Create a Slack Apps account.
2. Invite the account to a channel.
3. Get the token and signing secret from the Slack Apps.
4. Set secret SLACK_TOKEN and SLACK_SINGNING_SECRET(default `/notico/prod`) on AWS Secret Manager.
5. Set environment variables.
e.g.
```
export DOMAIN="..."
export NOTICE_CHANNEL_ID="CXXXXXXXX"
export LAMBROLL_TFSTATE="s3://..."
export TF_BACKEND_BUCKET="..."
export TF_BACKEND_KEY="..."
export TF_BACKEND_REGION="..."
```
6. Setup [aqua](jttps://aquaproj.github.io/) and `aqua i`
7. Setup infrastructures by terraform.
```sh
$ task terraform:init
$ task terraform:plan
$ task terraform:apply
```
8. Deploy the lambda function.
```sh
$ task lambda:deploy
```

## Supported Events

- channel_created
- channel_deleted
- channel_rename
- channel_archive
- channel_unarchive
- team_join
- subteam_created

## LICENSE

The MIT License (MIT)

Copyright (c) 2016-2024 KAYAC Inc.
