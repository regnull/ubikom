package newscache

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/translate"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"golang.org/x/text/language"
	"jaytaylor.com/html2text"
)

const (
	articleTTL = 24 * time.Hour
)

var keywords = []string{"ukrain", "russia", "moscow", "kiev", "kyiv", "putin", "lviv", "nato"}

type Entry struct {
	Url      string
	Added    time.Time
	Headline string
	Content  string
}

type Cache struct {
	entries map[int]*Entry
	lock    sync.RWMutex
	id      int
}

type Headline struct {
	Title string
	ID    int
}

func New() *Cache {
	return &Cache{
		entries: make(map[int]*Entry),
		id:      1000,
	}
}

func (c *Cache) GetHeadlines() []*Headline {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var ret []*Headline
	var ids []int
	for k := range c.entries {
		ids = append(ids, k)
	}
	sort.Ints(ids)
	// Go in the reverse order (newest articles first).
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		ret = append(ret, &Headline{Title: c.entries[id].Headline, ID: id})
	}
	return ret
}

func (c *Cache) GetArticle(id int) (string, string, error) {
	c.lock.RLock()
	e := c.entries[id]
	if e == nil {
		c.lock.RUnlock()
		return "", "", errors.New("not found")
	}
	if e.Content != "" {
		log.Debug().Int("id", id).Msg("article found in cache.")
		c.lock.RUnlock()
		return e.Headline, e.Content, nil
	}
	c.lock.RUnlock()

	c.lock.Lock()
	defer c.lock.Unlock()

	// Must check again because we released the lock.
	e = c.entries[id]
	if e == nil {
		return "", "", errors.New("not found")
	}
	if e.Content != "" {
		return e.Headline, e.Content, nil
	}

	log.Debug().Int("id", id).Msg("retrieving article")
	resp, err := http.Get(e.Url)
	if err != nil {
		return "", "", errors.New("failed to get article")
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	text, err := html2text.FromString(string(page), html2text.Options{PrettyTables: true})
	if err != nil {
		return "", "", err
	}

	log.Debug().Int("id", id).Msg("translating article")
	translatedText, err := translateText("ru", text)
	if err != nil {
		return "", "", err
	}
	e.Content = translatedText
	return e.Headline, translatedText, nil
}

func (c *Cache) Refresh() error {
	log.Debug().Msg("refreshing headlines")
	headlines, err := getHeadlines("https://lite.cnn.com/en", keywords)
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.deleteExpiredLocked()
	for k, w := range headlines {
		articleUrl := "https://lite.cnn.com" + w
		if !c.isArticleInCacheLocked(articleUrl) {
			ruHeadline, err := translateText("ru", k)
			if err != nil {
				log.Error().Err(err).Msg("failed to translate text")
				continue
			}
			c.addArticleLocked(ruHeadline, articleUrl)
		}
	}
	return nil
}

func (c *Cache) deleteExpiredLocked() {
	var toDelete []int
	for k, v := range c.entries {
		if time.Since(v.Added) > articleTTL {
			toDelete = append(toDelete, k)
		}
	}

	for _, id := range toDelete {
		delete(c.entries, id)
	}
}

func (c *Cache) isArticleInCacheLocked(url string) bool {
	found := false
	for _, v := range c.entries {
		if v.Url == url {
			found = true
			break
		}
	}
	return found
}

func (c *Cache) addArticleLocked(headline string, url string) {
	c.id += 1
	log.Debug().Int("id", c.id).Msg("adding new article")
	c.entries[c.id] = &Entry{
		Url:      url,
		Headline: headline,
		Added:    time.Now(),
	}
}

func getHeadlines(url string, keywords []string) (map[string]string, error) {
	ret := make(map[string]string)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tkn := html.NewTokenizer(strings.NewReader(string(page)))

	var inLink bool
	var link string
	for {

		tt := tkn.Next()

		switch {
		case tt == html.ErrorToken:
			return ret, nil

		case tt == html.StartTagToken:
			t := tkn.Token()
			if t.Data == "a" {
				inLink = true
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						link = attr.Val
					}
				}
			}

		case tt == html.EndTagToken:
			t := tkn.Token()
			if t.Data == "a" {
				inLink = false
			}

		case tt == html.TextToken:
			if inLink {
				t := tkn.Token()
				include := false
				for _, kw := range keywords {
					if strings.Contains(strings.ToLower(t.Data), kw) {
						include = true
					}
				}
				if include {
					ret[t.Data] = link
				}
			}
		}
	}
}

func translateText(targetLanguage, text string) (string, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	opts := translate.Options{
		Format: "text",
	}

	lines := strings.Split(text, "\n")
	var newLines []string
	for _, line := range lines {
		resp, err := client.Translate(ctx, []string{line}, lang, &opts)
		if err != nil {
			return "", fmt.Errorf("translate: %v", err)
		}
		if len(resp) == 0 {
			return "", fmt.Errorf("translate returned empty response to text: %s", text)
		}
		newLines = append(newLines, resp[0].Text)
	}
	return strings.Join(newLines, "\n"), nil
}
