package parsing

import (
	"testing"

	"github.com/denisdubovitskiy/feedparser/internal/config"
	"github.com/denisdubovitskiy/feedparser/internal/database"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source Source
		body   string
		want   []Article
	}{
		{
			// Ссылки, подобные тем, что на его сайте, встречаются нечасто.
			name: "https://research.swtch.com/",
			source: &database.Source{
				Name: "https://research.swtch.com/",
				URL:  "https://research.swtch.com/",
				Config: config.SourceConfig{
					ArticleSelector: "ul.toc li",
					TitleSelector:   "a",
					DetailSelector:  "a",
				},
			},
			want: []Article{
				{
					Title:     "Go Testing By Example",
					DetailURL: "https://research.swtch.com/testing",
				},
				{
					Title:     "Open Source Supply Chain Security at Google",
					DetailURL: "https://research.swtch.com/acmscored",
				},
			},
			body: `<html>
  <body>
    <a class="rss" href="/feed.atom">RSS</a>
    <div class="main">
    <ul class="toc">
      <!-- Будет пропущено - нет названия по селектору -->
      <li class="toc-head"><b>Table of Contents</b> (favorites in bold)</li>
      <li><a href="testing" class="">Go Testing By Example</a> <span class="toc-when">December 2023</span>
        <div class="toc-summary">
        The importance of testing, and twenty tips for writing good tests.
        </div>
      </li>
      <li><a href="acmscored" class="">Open Source Supply Chain Security at Google</a> <span class="toc-when">November 2023</span>
        <div class="toc-summary">
        A remote talk at ACM SCORED 2023
        </div>
      </li>
    </ul>
    </div>
  </body>
</html>`,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()

			// act
			got, err := parser.Parse(tc.source, tc.body)

			// assert
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
