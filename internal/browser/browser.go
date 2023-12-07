package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func Run(ctx context.Context) error {
	return chromedp.Run(ctx)
}

func FetchHTML(ctx context.Context, url string) (string, error) {
	var body string
	return body, chromedp.Run(
		ctx,
		fetch.Enable(),
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML(`html`, &body),
	)
}

func NewLocalContext() (context.Context, context.CancelFunc) {
	tempDir, err := os.MkdirTemp("", "chromedp")
	if err != nil {
		panic(fmt.Sprintf("browser-context: unable to create a temp directory: %s", err.Error()))
	}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(tempDir),
		// Headless,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	)

	allocCtx, cancelAllocCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	browserCtx, cancelBrowserCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	chromedp.ListenTarget(browserCtx, disableFetchExceptScripts(browserCtx))

	return browserCtx, func() {
		cancelAllocCtx()
		cancelBrowserCtx()
		_ = os.RemoveAll(tempDir)
	}
}

func NewRemoteContext(url string) (context.Context, context.CancelFunc) {
	allocCtx, cancelAllocCtx := chromedp.NewRemoteAllocator(context.Background(), url)
	browserCtx, cancelBrowserCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	chromedp.ListenTarget(browserCtx, disableFetchExceptScripts(browserCtx))

	return browserCtx, func() {
		cancelAllocCtx()
		cancelBrowserCtx()
	}
}

func disableFetchExceptScripts(ctx context.Context) func(event interface{}) {
	return func(event interface{}) {
		switch ev := event.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				// Картинки, шрифты, стили, медиа.
				if isResourceTypeMedia(ev.ResourceType) {
					if err := cancelRequest(ctx, ev.RequestID); err != nil {
						log.Printf(
							"browser: unable to cancel request for resource type %s: %v\n",
							ev.ResourceType.String(),
							err,
						)
					}
					return
				}

				if err := continueRequest(ctx, ev.RequestID); err != nil {
					log.Printf(
						"browser: unable to continue request for resource type %s: %v\n",
						ev.ResourceType.String(),
						err,
					)
				}
			}()
		}
	}
}

func cancelRequest(ctx context.Context, id fetch.RequestID) error {
	return fetch.FailRequest(id, network.ErrorReasonBlockedByClient).Do(ctx)
}

func continueRequest(ctx context.Context, id fetch.RequestID) error {
	return fetch.ContinueRequest(id).Do(ctx)
}

func isResourceTypeMedia(t network.ResourceType) bool {
	return t == network.ResourceTypeImage ||
		t == network.ResourceTypeStylesheet ||
		t == network.ResourceTypeFont ||
		t == network.ResourceTypeMedia ||
		t == network.ResourceTypeManifest
}
