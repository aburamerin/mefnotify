package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

func init() {
	beeep.DefaultDuration = 5000 // 5sec
}

const (
	URL            = "https://dofamin.org/index.php?v=Diary"
	ScrapeInterval = 30
)

type PostInfo struct {
	PostDate time.Time
	URL      string
}

type Post struct {
	ID       string
	Info     PostInfo
	Nickname string
	Text     string
}

func getIcon(s string) []byte {
	b, err := os.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func DofaminScrape() {
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

	doc.Find(".post").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Search last 3
		if i == 3 {
			return false
		}

		p := Post{}

		p.ID, _ = s.Attr("id")
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

		nickname := strings.TrimSpace(info.Find("div.post__author-wrapper").Find("a.post__author").Text())
		text := s.Find("div.post__content").Find("div.post__text").Text()
		preview := strings.Join(strings.Split(text, " ")[0:3], " ")

		fmt.Printf("%+v\n", p)
		SendNotify(nickname, preview, url)
		return true
	})
}

func SendNotify(nickname string, preview string, url string) {
	msg := preview
	if url != "" {
		msg = "<a href='" + url + "'>" + preview + " ...</a>"
	}
	embededLogo := "assets/info.png"

	err := beeep.Notify(nickname, msg, embededLogo)
	if err != nil {
		panic(err)
	}
}

func onExit() {
	SendNotify("Ael", "Пока", "")
}

func onReady() {
	systray.SetIcon(getIcon("assets/nf.ico"))

	go func() {
		for {
			DofaminScrape()
			time.Sleep(ScrapeInterval * time.Second)
		}
	}()
}

func main() {
	systray.Run(onReady, onExit)
}
