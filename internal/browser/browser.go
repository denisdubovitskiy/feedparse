package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/media"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func Run(ctx context.Context) error {
	return chromedp.Run(ctx)
}

func Headless(a *chromedp.ExecAllocator) {
	chromedp.Flag("headless", false)(a)
	// Like in Puppeteer.
	chromedp.Flag("hide-scrollbars", false)(a)
	chromedp.Flag("mute-audio", false)(a)
}

func FetchHTML(ctx context.Context, url string) (string, error) {
	var body string
	return body, chromedp.Run(
		ctx,
		fetch.Enable(),
		media.Disable(),
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML(`html`, &body),
	)
}

func NewContext() (context.Context, context.CancelFunc) {
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
	chromedp.ListenTarget(browserCtx, DisableFetchExceptScripts(browserCtx))

	return browserCtx, func() {
		cancelAllocCtx()
		cancelBrowserCtx()
		_ = os.RemoveAll(tempDir)
	}
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
