package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/slack-go/slack"
)

var (
	token   string
	channel string
	version string
	domain  string
)

func main() {
	var (
		showVersion bool
	)
	flag.StringVar(&channel, "channel", "#admins", "Channel to post notification message")
	flag.BoolVar(&showVersion, "version", false, "Show versrion")
	flag.Parse()
	if showVersion {
		fmt.Println("notico version", version)
		return
	}
	if token = os.Getenv("SLACK_TOKEN"); token == "" {
		fmt.Println("SLACK_TOKEN environment variable is not set.")
		os.Exit(1)
	}

	// set log level
	var minLevel logutils.LogLevel
	var debug bool
	if os.Getenv("DEBUG") != "" {
		minLevel = logutils.LogLevel("debug")
		debug = true
	} else {
		minLevel = logutils.LogLevel("info")
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: minLevel,
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	api := slack.New(token, slack.OptionDebug(debug))
	rtm := api.NewRTM()
	go rtm.ManageConnection()
Loop:
	for {
		var notifyMsg string
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ChannelCreatedEvent:
				notifyMsg = fmt.Sprintf("<@%s> が #%s を作成しました", ev.Channel.Creator, ev.Channel.Name)
			case *slack.ChannelDeletedEvent:
				notifyMsg = fmt.Sprintf("<#%s> が削除されました", ev.Channel)
			case *slack.ChannelRenameEvent:
				notifyMsg = fmt.Sprintf("<#%s> が #%s にリネームされました", ev.Channel.ID, ev.Channel.Name)
			case *slack.ChannelArchiveEvent:
				notifyMsg = fmt.Sprintf("<@%s> が <#%s> をアーカイブしました", ev.User, ev.Channel)
			case *slack.ChannelUnarchiveEvent:
				notifyMsg = fmt.Sprintf("<@%s> が <#%s> をアーカイブ解除しました", ev.User, ev.Channel)
			case *slack.SubteamCreatedEvent:
				notifyMsg = fmt.Sprintf("<@%s> がユーザグループ <!subteam^%s|%s> を作成しました: %s", ev.Subteam.CreatedBy, ev.Subteam.ID, ev.Subteam.Handle, ev.Subteam.Description)
			case *slack.TeamJoinEvent:
				accoutType := ""
				if ev.User.IsBot {
					accoutType = "bot"
				} else if ev.User.IsUltraRestricted {
					accoutType = "single channel guest"
				} else if ev.User.IsRestricted {
					accoutType = "multi channel guest"
				} else {
					accoutType = "normal"
				}
				notifyMsg = fmt.Sprintf("<@%s> (%s) がチームにjoinしました (%s)", ev.User.ID, ev.User.Profile.Email, accoutType)
			case *slack.BotAddedEvent:
				notifyMsg = fmt.Sprintf("bot %s が追加されました https://%s.slack.com/services/%s", ev.Bot.Name, domain, ev.Bot.ID)
			case *slack.ConnectedEvent:
				domain = ev.Info.Team.Domain
				log.Printf("[info] Team Info: %#v", ev.Info.Team)
			case *slack.InvalidAuthEvent:
				log.Printf("[error] Invalid credentials")
				break Loop
			default:
				// Ignore other events..
				log.Printf("[debug] Unexpected: %#v\n", msg.Data)
			}
		}
		if notifyMsg != "" {
			sendMessage(Message{
				Text:    notifyMsg,
				Channel: channel,
			})
			log.Println("[info] msg:", notifyMsg)
		}
	}
}

type Message struct {
	Text    string
	Channel string
}

func sendMessage(msg Message) {
	q := url.Values{
		"token":      {token},
		"channel":    {msg.Channel},
		"text":       {msg.Text},
		"link_names": {"1"},
		"as_user":    {"1"},
	}
	log.Println("[info] ", q.Encode())
	resp, err := http.PostForm("https://slack.com/api/chat.postMessage", q)
	if err != nil {
		log.Println("[warn] err response", err)
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[warn] readall failed", err)
	}
	log.Println("[info] ", string(s))
}
