package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/nlopes/slack"
)

var (
	token   string
	channel string
	version string
	domain  string
)

func main() {
	var showVersion bool
	flag.StringVar(&channel, "channel", "#admins", "Channel to post notification message")
	flag.BoolVar(&showVersion, "version", false, "Show versrion")
	if showVersion {
		fmt.Println("notico version", version)
		return
	}
	if token = os.Getenv("SLACK_TOKEN"); token == "" {
		fmt.Println("SLACK_TOKEN environment variable is not set.")
		os.Exit(1)
	}
	api := slack.New(token)
	if os.Getenv("DEBUG") != "" {
		api.SetDebug(true)
	}
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
			case *slack.TeamJoinEvent:
				accoutType := ""
				if ev.User.IsBot {
					accoutType = "bot"
				} else if ev.User.IsUltraRestricted {
					accoutType = "guest"
				} else if ev.User.IsRestricted {
					accoutType = "restricted"
				} else {
					accoutType = "normal"
				}
				notifyMsg = fmt.Sprintf("<@%s> (%s) がチームにjoinしました (%s)", ev.User.ID, ev.User.Profile.Email, accoutType)
			case *slack.BotAddedEvent:
				notifyMsg = fmt.Sprintf("bot %s が追加されました https://%s.slack.com/services/%s", ev.Bot.Name, domain, ev.Bot.ID)
			case *slack.ConnectedEvent:
				domain = ev.Info.Team.Domain
				log.Printf("Team Info: %#v", ev.Info.Team)
			case *slack.InvalidAuthEvent:
				log.Printf("Invalid credentials")
				break Loop
			default:
				// Ignore other events..
				log.Printf("Unexpected: %#v\n", msg.Data)
			}
		}
		if notifyMsg != "" {
			sendMessage(Message{
				Text:    notifyMsg,
				Channel: channel,
			})
			log.Println("msg:", notifyMsg)
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
	log.Println(q.Encode())
	resp, err := http.PostForm("https://slack.com/api/chat.postMessage", q)
	if err != nil {
		log.Println("err response", err)
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("readall failed", err)
	}
	log.Println(string(s))
}
