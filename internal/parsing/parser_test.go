package parsing

import (
	"testing"

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
			source: Source{
				Name:            "https://research.swtch.com/",
				URL:             "https://research.swtch.com/",
				ArticleSelector: "ul.toc li",
				TitleSelector:   "a",
				DetailSelector:  "a",
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
		{
			name: "https://eli.thegreenplace.net/tag/go",
			source: Source{
				Name:            "https://eli.thegreenplace.net/tag/go",
				URL:             "https://eli.thegreenplace.net/tag/go",
				ArticleSelector: "table.archive-list tbody tr td",
				TitleSelector:   "a",
				DetailSelector:  "a",
			},
			want: []Article{
				{
					Title:     "Using Ollama with LangChainGo",
					DetailURL: "https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/",
				},
				{
					Title:     "Retrieval Augmented Generation in Go",
					DetailURL: "https://eli.thegreenplace.net/2023/retrieval-augmented-generation-in-go/",
				},
				{
					Title:     "Better HTTP server routing in Go 1.22",
					DetailURL: "https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122/",
				},
			},
			body: `<!DOCTYPE html>
<html>
<head>
    <title>Articles in tag "Go"</title>
</head>
<body>
<div class="container">
    <div class="row">
        <section id="content">
            <h1>Articles in tag "Go"</h1>
            <table class="archive-list">
                <tr>
                    <td style="padding-right: 10px">2023.11.22:</td>
                    <td><a href='https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/'>Using Ollama with
                        LangChainGo</a></td>
                </tr>
                <tr>
                    <td style="padding-right: 10px">2023.11.10:</td>
                    <td><a href='https://eli.thegreenplace.net/2023/retrieval-augmented-generation-in-go/'>Retrieval
                        Augmented Generation in Go</a></td>
                </tr>
                <tr>
                    <td style="padding-right: 10px">2023.10.16:</td>
                    <td><a href='https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122/'>Better HTTP
                        server routing in Go 1.22</a></td>
                </tr>
            </table>
        </section>
    </div>
</div>
</body>
</html>
`,
		},
		{
			name: "https://threedots.tech/",
			source: Source{
				URL:             "https://threedots.tech/",
				Name:            "https://threedots.tech/",
				ArticleSelector: "article.post-entry",
				TitleSelector:   "header.post-header > h3.post-title a",
				DetailSelector:  "header.post-header > h3.post-title a",
			},
			want: []Article{
				{
					Title:     "Making Games in Go for Absolute Beginners",
					DetailURL: "https://threedots.tech/post/making-games-in-go/",
				},
				{
					Title:     "Watermill 1.3 released, an open-source event-driven Go library",
					DetailURL: "https://threedots.tech/post/watermill-1-3/",
				},
			},
			body: `<html lang=en>
<body>
<section class="main post-list">
    <article class=post-entry>
        <header class=post-header>
            <h3 class=post-title>
                <a href=https://threedots.tech/post/making-games-in-go/ class=post-link>Making Games in Go for Absolute Beginners</a></h3>
            <p class=post-meta>
                <a href=https://twitter.com/m1_10sz target=_blank>
                    <img src="https://gravatar.com/avatar/4b7742aef224a14ed9de437b55057033?s=160" class=author-avatar
                         alt=Author>
                </a>
                @Miłosz Smółka · Nov 24, 2023
            </p>
        </header>
        <img class=post-cover src=https://threedots.tech/post/making-games-in-go/cover.png
             alt="Making Games in Go for Absolute Beginners">
        <p class=post-summary>Here’s a rant I often see in developer communities:
            I used to love programming because I like building stuff. But my full-time job killed my passion. I spend
            more time in meetings, fighting over deadlines, and arguing in reviews than working with code. Am I burned
            out? Is there hope, or do I need a new hobby?
            Sounds familiar? No wonder we keep looking forward to using a new framework or database — we’re bored.</p>
        <footer class=post-footer>
            <a class=read-more href=https://threedots.tech/post/making-games-in-go/>Read More →</a>
        </footer>
    </article>
    <article class=post-entry>
        <header class=post-header>
            <h3 class=post-title><a href=https://threedots.tech/post/watermill-1-3/ class=post-link>Watermill 1.3
                released, an open-source event-driven Go library</a></h3>
            <p class=post-meta>
                <a href=https://twitter.com/m1_10sz target=_blank>
                    <img src="https://gravatar.com/avatar/4b7742aef224a14ed9de437b55057033?s=160" class=author-avatar
                         alt=Author>
                </a>
                @Miłosz Smółka · Sep 25, 2023
            </p>
        </header>
        <img class=post-cover src=https://threedots.tech/post/watermill-1-3/cover.png
             alt="Watermill 1.3 released, an open-source event-driven Go library">
        <p class=post-summary>Hey, it’s been a long time!
            We’re happy to share that Watermill v1.3 is now out!
            What is Watermill Watermill is an open-source library for building message-driven or event-driven
            applications the easy way in Go. Our definition of “easy” is as easy as building an HTTP server in Go. With
            all that, it’s a library, not a framework. So your application is not tied to Watermill forever.
            Currently, Watermill has over 6k stars on GitHub, has over 50 contributors, and has been used by numerous
            projects in the last 4 years.</p>
        <footer class=post-footer>
            <a class=read-more href=https://threedots.tech/post/watermill-1-3/>Read More →</a>
        </footer>
    </article>
</section>
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
