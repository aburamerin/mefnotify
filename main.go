package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mefnotify/pkg/posts"
	"mefnotify/pkg/telegram"

	"github.com/PuerkitoBio/goquery"
)

const (
	PreviewLen     = 64
	ScrapeInterval = 60
	URL            = "https://dofamin.org/index.php?v=Diary"
)

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length]) + "..."
}

func getIcon(s string) []byte {
	b, err := os.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func DofaminScrape(db *sql.DB) (sliceOfPosts posts.PostsSlice) {
	// Request the HTML page.
	res, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	sliceOfPosts = posts.PostsSlice{}

	doc.Find(".post").EachWithBreak(func(i int, s *goquery.Selection) bool {
		p := posts.Post{}
		id, ok := s.Attr("id")
		if ok {
			id = strings.TrimPrefix(id, "postid_")
			p.ID, _ = strconv.ParseInt(id, 10, 64)
		}
		// style, _ := s.Attr("style")
		info := s.Find("div.post__info")

		url, ok := info.Find("a.post__option").First().Attr("href")
		if !ok {
			url = ""
		}
		p.Info.URL = url

		p.Info.PostDate, err = time.Parse("02.01.2006 15:04", info.Find("div.post__date").Text())
		if err != nil {
			p.Info.PostDate = time.Time{}
		}

		p.Author = strings.TrimSpace(info.Find("div.post__author-wrapper").Find("a.post__author").Text())
		p.Content = s.Find("div.post__content").Find("div.post__text").First().Text()
		p.Preview = TruncateString(p.Content, PreviewLen)

		// search in database
		if !posts.FindPost(db, p.ID) {
			log.Printf("Post %d not found, add to slice for sending.", p.ID)
			sliceOfPosts = append(sliceOfPosts, p)
		}

		return true
	})

	return sliceOfPosts
}

func SendToChat(db *sql.DB, client *telegram.Client, sliceOfPosts posts.PostsSlice) {
	// if no new post - skipping
	if len(sliceOfPosts) == 0 {
		return
	}

	// sort by date
	sort.Slice(sliceOfPosts, func(i, j int) bool {
		return sliceOfPosts[i].Info.PostDate.Before(sliceOfPosts[j].Info.PostDate)
	})

	// send here and save to db
	for _, p := range sliceOfPosts {
		err := client.SendMessage(p)
		if err == nil {
			err := posts.StorePost(db, p)
			if err != nil {
				// Fail if something wrong with database
				log.Fatalf("ERROR storing post %s: %s", p.String(), err)
			}
		} else {
			log.Printf("ERROR sending %d message %s: %s", p.ID, p.Preview, err)
		}
	}
}

func main() {
	chatID, err := strconv.Atoi(os.Getenv("MEF_CHATID"))
	if err != nil {
		log.Fatalf("FATAL: wrong telegram chat_id value: %s", err)
	}

	token := os.Getenv("MEF_TGTOKEN")
	client, err := telegram.New(token, int64(chatID))
	if err != nil {
		log.Fatalf("FATAL: can'c connect telegram bot: %s", err)
	}

	DSN := os.Getenv("MEF_DSN")
	db, err := posts.NewDB(DSN)
	if err != nil {
		log.Fatalf("FATAL: Can't connect to db: %s", err)
	}

	// scrape in endless loop
	go func() {
		for {
			allPosts := DofaminScrape(db)
			SendToChat(db, client, allPosts)
			time.Sleep(ScrapeInterval * time.Second)
		}
	}()

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	// time for cleanup before exit
	log.Println("Bye!")
}
