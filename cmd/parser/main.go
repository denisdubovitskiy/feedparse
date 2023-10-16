package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denisdubovitskiy/feedparser/internal/browser"
	"github.com/denisdubovitskiy/feedparser/internal/database"
	"github.com/denisdubovitskiy/feedparser/internal/parsing"
	"github.com/denisdubovitskiy/feedparser/internal/task"
	"github.com/denisdubovitskiy/feedparser/internal/telegram"
	"github.com/denisdubovitskiy/feedparser/internal/unix"
)

var (
	databasePath string
)

func init() {
	flag.StringVar(&databasePath, "database", "", "database filename")
	flag.Parse()
}

var (
	token              = os.Getenv("CRAWLER_TG_TOKEN")
	channel            = os.Getenv("CRAWLER_TG_CHANNEL")
	browserURL         = os.Getenv("CRAWLER_BROWSER_URL")
	browserLocation    = os.Getenv("CRAWLER_BROWSER_LOCATION")
	isPublisherEnabled = len(os.Getenv("CRAWLER_PUBLISHER_ENABLED")) > 0
)

func main() {
	appCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	db, err := database.New(databasePath)
	if err != nil {
		log.Fatalln(err)
	}

	service := database.NewService(db)

	var browserCtx context.Context
	var browserCtxCancel context.CancelFunc
	if browserLocation == "remote" {
		browserCtx, browserCtxCancel = browser.NewRemoteContext(browserURL)
		defer browserCtxCancel()
	} else {
		browserCtx, browserCtxCancel = browser.NewLocalContext()
		defer browserCtxCancel()
	}

	if err := browser.Run(browserCtx); err != nil {
		log.Fatal(err)
	}

	parser := parsing.NewParser()
	runner := task.NewRunner(service, 3)

	crawlTicker := time.NewTicker(time.Minute)
	crawl := func() {
		runnerErr := runner.ForEachSource(context.Background(), func(source *database.Source) error {
			log.Printf("source: %s requesting", source.String())

			ctx, cancel := context.WithTimeout(browserCtx, 10*time.Second)
			defer cancel()

			body, err := browser.FetchHTML(ctx, source.URL)
			if err != nil {
				log.Printf("source: %s request failed", source.String())
				return err
			}

			log.Printf("source: %s request succeded", source.String())

			articles, err := parser.Parse(source, body)
			if err != nil {
				return err
			}

			for _, article := range articles {
				saveArticleErr := service.SaveArticle(context.Background(), database.SaveArticleParams{
					SourceID: source.ID,
					Title:    article.Title,
					Url:      article.DetailURL,
					Added:    unix.TimeNow(),
				})
				if saveArticleErr != nil {
					log.Printf("source: %s unable to save: %v", article.String(), saveArticleErr)
					continue
				}

				log.Printf("source: %s %s saved", source.String(), article.String())
			}

			return nil
		})

		if runnerErr != nil {
			log.Println(runnerErr.Error())
		}
	}

	go func() {
		defer crawlTicker.Stop()

		crawl()

		for {
			select {
			case <-crawlTicker.C:
				log.Println("crawler: tick")
				crawl()
			case <-appCtx.Done():
				return
			}
		}
	}()

	if isPublisherEnabled {
		publisher := telegram.NewPublisher(token, channel)
		sendTicker := time.NewTicker(2 * time.Second)

		go func() {
			defer sendTicker.Stop()

			sendAfter := time.Now()

			for {
				select {

				case <-sendTicker.C:
					log.Println("sender: tick")

					if sendAfter.After(time.Now()) {
						continue
					}

					sendErr := service.SelectUnsent(context.Background(), func(article database.Article) error {
						log.Printf("sender: sending article %s", article.String())

						if err := publisher.PublishPost(article.Source, article.Title, article.URL); err != nil {
							if after, ok := telegram.CanRetry(err); ok {
								sendAfter = time.Now().Add(time.Duration(after) * time.Second)
								log.Printf("sender: rate limit exceeded, retrying after %d seconds", after)
								return err
							}
							log.Printf("sender: unable to send article %s", article.String())
							return err
						}

						log.Printf("sender: article sent %s", article.String())

						return nil
					})
					if sendErr != nil {
						if errors.Is(sendErr, sql.ErrNoRows) {
							log.Println("sender: no unsent articles found")
							continue
						}
						if _, ok := telegram.CanRetry(err); ok {
							continue
						}
						log.Printf("sender: unable to send an article: %v", err)
					}
				case <-appCtx.Done():
					return
				}
			}
		}()
	}

	<-appCtx.Done()
}
