package notico

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Notico struct {
	client          *slack.Client
	domain          string
	noticeChannelID string
	signingSecret   string
}

func New(token, signingSecret, domain, noticeChannelID string) *Notico {
	return &Notico{
		client:          slack.New(token),
		domain:          domain,
		noticeChannelID: noticeChannelID,
		signingSecret:   signingSecret,
	}
}

func (n *Notico) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithoutCancel(r.Context())
	if r.Header.Get("X-Slack-Retry-Num") != "" && r.Header.Get("X-Slack-Retry-Reason") == "http_timeout" {
		slog.WarnContext(ctx, "skip retry", slog.Any("header", r.Header))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to read body: %w", err))
		return
	}
	if err := r.Body.Close(); err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to close request body: %w", err))
		return
	}

	sv, err := slack.NewSecretsVerifier(r.Header, n.signingSecret)
	if err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to create secrets verifier: %w", err))
		return
	}
	if _, err := sv.Write(buf); err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to write request body: %w", err))
		return
	}
	if err := sv.Ensure(); err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to verify request: %w", err))
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(buf), slackevents.OptionNoVerifyToken())
	if err != nil {
		n.handleError(ctx, w, fmt.Errorf("failed to parse event: %w", err))
		return
	}
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		if err := json.Unmarshal(buf, &r); err != nil {
			n.handleError(ctx, w, fmt.Errorf("failed to unmarshal challenge response: %w", err))
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		return
	}
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		event := eventsAPIEvent.InnerEvent
		if err := n.reaction(r.Context(), event); err != nil {
			n.handleError(ctx, w, fmt.Errorf("failed to reaction: %w", err))
			return
		}
	}
}

func (n *Notico) handleError(ctx context.Context, w http.ResponseWriter, err error) {
	slog.ErrorContext(ctx, "error", slog.Any("error", err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func (n *Notico) reaction(ctx context.Context, event slackevents.EventsAPIInnerEvent) error {
	switch ev := event.Data.(type) {
	case *slackevents.ChannelCreatedEvent:
		notifyMsg := fmt.Sprintf("<@%s> が #%s を作成しました", ev.Channel.Creator, ev.Channel.Name)
		slog.InfoContext(ctx, "channel created", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.ChannelDeletedEvent:
		notifyMsg := fmt.Sprintf("<#%s> が削除されました", ev.Channel)
		slog.InfoContext(ctx, "channel deleted", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.ChannelRenameEvent:
		notifyMsg := fmt.Sprintf("<#%s> が #%s にリネームされました", ev.Channel.ID, ev.Channel.Name)
		slog.InfoContext(ctx, "channel rename", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.ChannelArchiveEvent:
		notifyMsg := fmt.Sprintf("<@%s> が <#%s> をアーカイブしました", ev.User, ev.Channel)
		slog.InfoContext(ctx, "channel archive", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.ChannelUnarchiveEvent:
		notifyMsg := fmt.Sprintf("<@%s> が <#%s> をアーカイブ解除しました", ev.User, ev.Channel)
		slog.InfoContext(ctx, "channel unarchive", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.SubteamCreatedEvent:
		notifyMsg := fmt.Sprintf("<@%s> がユーザグループ <!subteam^%s|%s> を作成しました: %s", ev.Subteam.CreatedBy, ev.Subteam.ID, ev.Subteam.Handle, ev.Subteam.Description)
		slog.InfoContext(ctx, "subteam created", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slackevents.TeamJoinEvent:
		var accountType string
		if ev.User.IsBot {
			accountType = "bot"
		} else if ev.User.IsUltraRestricted {
			accountType = "single channel guest"
		} else if ev.User.IsRestricted {
			accountType = "multi channel guest"
		} else {
			accountType = "normal"
		}
		notifyMsg := fmt.Sprintf("<@%s> (%s) がチームにjoinしました (%s)", ev.User.ID, ev.User.Profile.Email, accountType)
		slog.InfoContext(ctx, "team join", slog.String("message", notifyMsg))
		if err := n.sendMessage(ctx, notifyMsg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case *slack.MessageEvent:
		slog.InfoContext(ctx, "message text", slog.String("text", ev.Text))
		if strings.HasPrefix(ev.Text, "notico:") && strings.Contains(ev.Text, "ping") {
			slog.InfoContext(ctx, "message channel", slog.String("channel", ev.Channel))
			if err := n.sendMessage(ctx, "pong"); err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
		}
	}
	return nil
}

func (n *Notico) sendMessage(ctx context.Context, message string) error {
	param := slack.NewPostMessageParameters()
	param.LinkNames = 1
	if _, _, _, err := n.client.SendMessageContext(
		ctx,
		n.noticeChannelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionPostMessageParameters(param),
	); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
