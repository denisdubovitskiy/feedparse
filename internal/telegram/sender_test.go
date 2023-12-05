package telegram

import (
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mk struct {
	client    *ClientMock
	publisher *Publisher
	channel   string
}

func newMk(t *testing.T) (*mk, func()) {
	mc := minimock.NewController(t)
	client := NewClientMock(mc)
	channel := "test-channel"
	m := &mk{
		channel:   channel,
		client:    client,
		publisher: NewWithClient(client, channel),
	}
	return m, mc.Finish
}

type testArticle struct {
	source   string
	title    string
	url      string
	channels []string
	tags     []string
}

var (
	emptyChannels    []string
	emptyTags        []string
	multipleChannels = []string{
		"chan1",
		"chan2",
	}
	nilMessage tgbotapi.Message
	nilError   error
)

func testNewArticle(channels, tags []string) testArticle {
	return testArticle{
		source:   "test-source",
		title:    "test-title",
		url:      "test-url",
		channels: channels,
		tags:     tags,
	}
}

func TestSender(t *testing.T) {

	t.Parallel()

	t.Run("single channel", func(t *testing.T) {
		t.Parallel()

		m, cleanup := newMk(t)
		defer cleanup()

		article := testNewArticle(emptyChannels, emptyTags)
		text := formatMessage(article.source, article.title, article.url, article.tags)
		wantMessage := newMarkdownMessage(m.publisher.channel, text)

		m.client.SendMock.
			Expect(wantMessage).
			Return(nilMessage, nilError)

		// act
		err := m.publisher.PublishPost(
			article.source,
			article.title,
			article.url,
			article.channels,
			article.tags,
		)

		// assert
		require.NoError(t, err)
	})

	t.Run("multiple channels", func(t *testing.T) {
		t.Parallel()

		m, cleanup := newMk(t)
		defer cleanup()

		article := testNewArticle(multipleChannels, emptyTags)
		text := formatMessage(article.source, article.title, article.url, article.tags)

		wantMessage1 := newMarkdownMessage(article.channels[0], text)
		wantMessage2 := newMarkdownMessage(article.channels[1], text)

		m.client.SendMock.When(wantMessage1).Then(nilMessage, nilError)
		m.client.SendMock.When(wantMessage2).Then(nilMessage, nilError)

		// act
		err := m.publisher.PublishPost(
			article.source,
			article.title,
			article.url,
			article.channels,
			article.tags,
		)

		// assert
		require.NoError(t, err)
	})

	t.Run("generic error", func(t *testing.T) {
		t.Parallel()

		m, cleanup := newMk(t)
		defer cleanup()

		article := testNewArticle(emptyChannels, emptyTags)
		text := formatMessage(article.source, article.title, article.url, article.tags)
		wantMessage := newMarkdownMessage(m.publisher.channel, text)

		m.client.SendMock.Expect(wantMessage).Return(nilMessage, assert.AnError)

		// act
		err := m.publisher.PublishPost(
			article.source,
			article.title,
			article.url,
			article.channels,
			article.tags,
		)

		// assert
		require.Error(t, err)
		require.ErrorAs(t, assert.AnError, &err)
	})
}
