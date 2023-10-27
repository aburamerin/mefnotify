package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
	"mefnotify/pkg/posts"
	"mefnotify/pkg/telegram"

	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

const PREVIEW_LEN = 32

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length]) + "..."
}

func init() {
	beeep.DefaultDuration = 5000 // 5sec
}

const (
	URL            = "https://dofamin.org/index.php?v=Diary"
	ScrapeInterval = 30
)

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
		p.Content = s.Find("div.post__content").Find("div.post__text").Text()
		p.Preview = TruncateString(p.Content, PREVIEW_LEN)

		fmt.Printf("%+v\n", p)

		if posts.FindPost(db, p.ID) {
			log.Printf("Post found, skipping.")
		} else {
			log.Printf("Post not found, sending.")
			sliceOfPosts = append(sliceOfPosts, p)
		}

		return true
	})

	return sliceOfPosts
}

func SendToChat(db *sql.DB, client *telegram.Client, sliceOfPosts posts.PostsSlice) {
	if len(sliceOfPosts) == 0 {
		log.Printf("No new posts.")
		return
	}

	// sort by date
	sort.Slice(sliceOfPosts, func(i, j int) bool {
		return sliceOfPosts[i].Info.PostDate.Before(sliceOfPosts[j].Info.PostDate)
	})

	// send here
	for _, p := range sliceOfPosts {
		err := client.SendMessage(p)
		if err == nil {
			posts.StorePost(db, p)
		}
	}
}

func onExit() {
	log.Println("Bye.")
}

func onReady() {
	systray.SetIcon(getIcon("assets/nf.ico"))

	chatID, _ := strconv.Atoi(os.Getenv("MEF_CHATID"))

	token := os.Getenv("MEF_TGTOKEN")
	client := telegram.New(token, int64(chatID))
	db, _ := posts.NewDB("./posts.db")

	go func() {
		for {
			allPosts := DofaminScrape(db)
			SendToChat(db, client, allPosts)
			time.Sleep(ScrapeInterval * time.Second)
		}
	}()
}

func main() {
	systray.Run(onReady, onExit)
}
