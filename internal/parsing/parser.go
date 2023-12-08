package parsing

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/denisdubovitskiy/feedparser/internal/database"
)

type Source = *database.Source

type Article struct {
	Title     string
	DetailURL string
}

func (a Article) String() string {
	return fmt.Sprintf("Article(title=%s, url=%s)", a.Title, a.DetailURL)
}

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(source Source, body string) ([]Article, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parser: unable to parse %s: %v", source.String(), err)
	}

	log.Printf("parser: %s parsing succeeded", source.String())

	articles := make([]Article, 0, 10)

	doc.Find(source.Config.ArticleSelector).Each(func(i int, articleCard *goquery.Selection) {
		log.Printf("parser: %s parsing article %d", source.String(), i)

		title := articleCard.Find(source.Config.TitleSelector).Text()
		title = formatTitle(title)

		detailURL, _ := articleCard.Find(source.Config.DetailSelector).Attr("href")
		detailURL = strings.TrimSpace(detailURL)

		if len(title) == 0 {
			log.Printf("parser: %s article %d parsing failed - empty title", source.String(), i)
			return
		}

		if len(detailURL) == 0 {
			log.Printf("parser: %s article %d parsing failed - empty detail url", source.String(), i)
			return
		}

		if isLinkRelative(detailURL) {
			u, err := url.Parse(source.URL)
			if err != nil {
				return
			}

			detailURL = strings.TrimPrefix(detailURL, ".")
			detailURL = strings.TrimPrefix(detailURL, "/")
			detailURL = fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, detailURL)
		}

		article := Article{
			Title:     title,
			DetailURL: detailURL,
		}

		articles = append(articles, article)

		log.Printf("parser: %s parsing %s", source.String(), article.String())
	})

	return articles, nil
}

var regexpWhitespace = regexp.MustCompile(`\s+`)

func formatTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = regexpWhitespace.ReplaceAllLiteralString(s, " ")
	return s
}

func isLinkRelative(detailURL string) bool {
	if strings.HasPrefix(detailURL, "http://") ||
		strings.HasPrefix(detailURL, "https://") {
		return false
	}
	return true
}
