package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fujiwara/ridge"
	"github.com/kayac/notico"
	awsenvsecrets "github.com/mackee/envsecrets/dist/aws"
	sloghttp "github.com/samber/slog-http"
)

type CLI struct {
	SlackToken      string `help:"Slack API token" name:"slack-token" env:"SLACK_TOKEN" required:""`
	NoticeChannelID string `help:"Slack channel ID to post notification message" name:"notice-channel-id" env:"NOTICE_CHANNEL_ID" required:""`
	Domain          string `help:"Slack domain" name:"domain" env:"DOMAIN" required:""`
	SigningSecret   string `help:"Slack signing secret" name:"signing-secret" env:"SLACK_SIGNING_SECRET" required:""`
	LocalAddress    string `help:"Local address to listen" name:"local-address" default:":8080"`
}

func main() {
	ctx := context.Background()
	if err := awsenvsecrets.Load(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to load secrets", slog.Any("error", err))
		os.Exit(1)
	}

	var cli CLI
	kong.Parse(&cli)

	n := notico.New(cli.SlackToken, cli.SigningSecret, cli.Domain, cli.NoticeChannelID)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", n.Handler)
	handler := sloghttp.Recovery(mux)
	handler = sloghttp.New(slog.Default())(handler)

	ridge.Run(cli.LocalAddress, "/", handler)
}
