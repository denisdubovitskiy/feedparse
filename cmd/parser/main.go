package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/media"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/denisdubovitskiy/feedparser/internal/database"
	"github.com/denisdubovitskiy/feedparser/internal/unix"
)

var (
	databasePath string
)

func init() {
	flag.StringVar(&databasePath, "database", "", "database filename")
	flag.Parse()
}

func Headless(a *chromedp.ExecAllocator) {
	chromedp.Flag("headless", false)(a)
	// Like in Puppeteer.
	chromedp.Flag("hide-scrollbars", false)(a)
	chromedp.Flag("mute-audio", false)(a)
}

func main() {
	db, err := database.Open(databasePath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatalln(err)
	}

	service := database.NewService(db)

	dir, err := os.MkdirTemp("", "chromedp-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(dir),
		Headless,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	chromedp.ListenTarget(browserCtx, DisableFetchExceptScripts(browserCtx))
	defer cancel()

	if err := chromedp.Run(browserCtx); err != nil {
		log.Fatal(err)
	}

	runner := NewRunner(service, 3)
	runner.ForEachSource(func(source *database.Source) error {
		slog.Info(fmt.Sprintf("parser: requesting source %s", source.Name))

		var body string

		ctx, cancel := context.WithTimeout(browserCtx, 10*time.Second)
		defer cancel()

		chromeErr := chromedp.Run(
			ctx,
			fetch.Enable(),
			media.Disable(),
			chromedp.Navigate(source.URL),
			chromedp.Sleep(time.Second),
			chromedp.InnerHTML(`html`, &body),
		)
		if chromeErr != nil {
			return chromeErr
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		if err != nil {
			slog.Info(fmt.Sprintf("parser: source %s parsing failed", source.Name))
			return err
		}

		slog.Info(fmt.Sprintf("parser: source %s parsing succeeded", source.Name))

		doc.Find(source.Config.ArticleCard.Selector).Each(func(i int, articleCard *goquery.Selection) {
			slog.Info(fmt.Sprintf("parser: source %s parsing article %d", source.Name, i))

			title := articleCard.Find(source.Config.ArticleCard.Title.Selector).Text()
			title = strings.TrimSpace(title)

			detailURL, _ := articleCard.Find(source.Config.ArticleCard.Detail.Selector).Attr("href")
			detailURL = strings.TrimSpace(detailURL)

			if len(title) == 0 || len(detailURL) == 0 {
				slog.Error(fmt.Sprintf("parser: source %s article %d parsing failed - empty title or detail url", source.Name, i))
				return
			}

			// Относительные ссылки
			if strings.HasPrefix(detailURL, "./") ||
				strings.HasPrefix(detailURL, "/") {

				u, err := url.Parse(source.URL)
				if err != nil {
					return
				}

				detailURL = strings.TrimPrefix(detailURL, ".")
				detailURL = strings.TrimPrefix(detailURL, "/")
				detailURL = fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, detailURL)
			}

			slog.Info(fmt.Sprintf("parser: source %s parsing article %s (%s) succeeded", source.Name, title, detailURL))

			saveArticleErr := service.SaveArticle(context.Background(), database.SaveArticleParams{
				SourceID: source.ID,
				Title:    title,
				Url:      detailURL,
				Added:    unix.TimeNow(),
			})
			if saveArticleErr != nil {
				slog.Error(fmt.Sprintf("parser: unable to save article %s: %v", detailURL, saveArticleErr))
				return
			}

			slog.Error(fmt.Sprintf("parser: source %s article %s (%s) saved", source.Name, title, detailURL))
		})

		return nil
	})
}

func NewRunner(service *database.Service, maxRetries int64) *Runner {
	return &Runner{service: service, maxRetries: maxRetries}
}

type Runner struct {
	service    *database.Service
	maxRetries int64
}

func (r *Runner) ForEachSource(f func(source *database.Source) error) {
	if err := r.service.ResetRetries(context.Background()); err != nil {
		slog.Error(fmt.Sprintf("runner: unable to reset retries: %v", err))
		return
	}

	lastTimestamp, err := r.service.LastTimestamp(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			lastTimestamp = unix.TimeNow()
			if err := r.service.SetLastTimestamp(context.Background(), lastTimestamp); err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Fatalln(err)
		}
	}

	if lastTimestamp == 0 {
		lastTimestamp = unix.TimeNow()
		if err := r.service.SetLastTimestamp(context.Background(), lastTimestamp); err != nil {
			log.Fatalln(err)
		}
	}

	started := time.Now()

	slog.Info(fmt.Sprintf("runner: starting at %s", started.Format(time.DateTime)))

	unixStarted := lastTimestamp

	defer func() {
		if err := r.service.SetLastTimestamp(context.Background(), unix.TimeNow()); err != nil {
			log.Fatalln(err)
		}
	}()

	for {
		// Забираем из базы по одному источнику из тех, чья дата последнего
		// визита меньше, чем дата предыдущего запуска раннера.
		source, err := r.service.FetchOne(context.Background(), unixStarted, r.maxRetries)
		if err != nil {
			// Все источники пройдены.
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			slog.Error("runner: unable to fetch a source", slog.String("error", err.Error()))
			continue
		}

		if err := f(source); err != nil {
			slog.Error(fmt.Sprintf("runner: failed to process a source %s", source.URL))

			retriesUpdateErr := r.service.UpdateRetries(context.Background(), source.ID)
			if retriesUpdateErr != nil {
				slog.Error(fmt.Sprintf("runner: failed to update retries %s: %v", source.URL, retriesUpdateErr))
				return // TODO
			}

			continue
		}

		updateErr := r.service.UpdateLastVisited(context.Background(), source.ID, unix.TimeNow())
		if updateErr != nil {
			slog.Error(fmt.Sprintf("runner: %s update error: %v", source.URL, updateErr.Error()))
			return
		}
		slog.Info(fmt.Sprintf("runner: %s job finished", source.URL))
	}

	slog.Info(fmt.Sprintf("runner: finished, time taken: %f", time.Since(started).Seconds()))
}

func DisableFetchExceptScripts(ctx context.Context) func(event interface{}) {
	return func(event interface{}) {
		switch ev := event.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if ev.ResourceType == network.ResourceTypeImage ||
					ev.ResourceType == network.ResourceTypeStylesheet ||
					ev.ResourceType == network.ResourceTypeFont ||
					ev.ResourceType == network.ResourceTypeMedia ||
					ev.ResourceType == network.ResourceTypeManifest {
					fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				} else {
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}
			}()
		}
	}
}
