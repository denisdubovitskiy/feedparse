package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

func NewWithClient(client Client, channel string) *Publisher {
	channel = strings.TrimPrefix(channel, "@")
	channel = "@" + channel
	return &Publisher{
		client:  client,
		channel: channel,
	}
}

func NewPublisher(token, channel string) *Publisher {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize telegram client: %v", err))
	}

	return NewWithClient(client, channel)
}

type Publisher struct {
	client  Client
	channel string
}

func (p *Publisher) PublishPost(source, title, url string, channels, tags []string) error {

	var channelsToSend []string

	if len(channels) == 0 {
		channelsToSend = append(channelsToSend, p.channel)
	} else {
		channelsToSend = channels
	}

	text := formatMessage(source, title, url, tags)

	for _, channel := range channelsToSend {
		msg := newMarkdownMessage(channel, text)

		if _, err := p.client.Send(msg); err != nil {
			if e, ok := err.(*tgbotapi.Error); ok {
				if e.RetryAfter > 0 {
					return retryError(e, e.RetryAfter)
				}
			}

			return fmt.Errorf("publisher: unable to send a message: %w", err)
		}
	}

	return nil
}

func newMarkdownMessage(channel, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessageToChannel(channel, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	return msg
}

type RetryError struct {
	err   error
	after int
}

func (r RetryError) Error() string {
	return r.err.Error()
}

func CanRetry(err error) (int, bool) {
	if e, ok := err.(RetryError); ok {
		return e.after, true
	}
	return 0, false
}

var _ error = RetryError{}

func retryError(err error, after int) error {
	return RetryError{err: err, after: after}
}

const (
	templateBase     = `%s: [%s](%s)`
	templateWithTags = `%s: [%s](%s)
%s
`
)

func formatMessage(source, title, url string, tags []string) string {
	if len(tags) == 0 {
		return fmt.Sprintf(templateBase, source, title, url)
	}

	return fmt.Sprintf(templateWithTags, source, title, url, formatTags(tags))
}

func formatTags(tags []string) string {
	tagsWithHashtag := make([]string, len(tags))
	for i, t := range tags {
		if !strings.HasPrefix(t, "#") {
			t = "#" + t
		}
		tagsWithHashtag[i] = t
	}

	return strings.Join(tagsWithHashtag, ", ")
}
