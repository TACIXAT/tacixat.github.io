package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/gorilla/feeds"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type MetaData struct {
	Title       string
	Description string
}

type IndexData struct {
	Meta  MetaData
	Posts []PostData
}

type PostData struct {
	Meta      MetaData
	Tags      []string
	Date      time.Time
	Raw       template.HTML
	Published bool
}

type AboutData struct {
	Meta MetaData
	Raw  template.HTML
}

// Mon Jan 2 15:04:05 -0700 MST 2006
func (pd PostData) Datef() string {
	return pd.Date.Format("2006/01/02 15:04")
}

func writeTemplate(t []string, o string, d interface{}) error {
	it, err := template.ParseFiles(t...)
	if err != nil {
		return err
	}

	f, err := os.Create(o)
	if err != nil {
		return err
	}
	defer f.Close()

	return it.ExecuteTemplate(f, "base.gohtml", d)
}

func windowsBad(b []byte) []byte {
	return bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
}

func genPosts(id *IndexData) {
	files, err := ioutil.ReadDir("_posts")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		log.Println(f.Name())
		md, err := ioutil.ReadFile("_posts/" + f.Name())
		if err != nil {
			log.Println(err)
			continue
		}

		md = windowsBad(md)

		s := bytes.Split(md, []byte("\n---\n"))
		if len(s) < 2 {
			log.Println("Missing post metadata.")
			continue
		}
		md = bytes.Join(s[1:], []byte("\n---\n"))

		var pd PostData
		_, err = toml.Decode(string(s[0]), &pd)
		if err != nil {
			log.Println(err)
			continue
		}

		cr := NewChromaRenderer(
			ChromaOptions(html.TabWidth(4)))
		pd.Raw = template.HTML(blackfriday.Run(
			md, blackfriday.WithRenderer(cr)))
		writeTemplate(
			[]string{
				"_templates/post.gohtml",
				"_templates/base.gohtml",
			}, fmt.Sprintf("posts/%s.html", pd.Meta.Title), pd)
		id.Posts = append([]PostData{pd}, id.Posts...)
	}
}

func genAbout() {
	md, err := ioutil.ReadFile("_pages/about.md")
	if err != nil {
		log.Println(err)
		return
	}

	md = windowsBad(md)

	ad := AboutData{
		Meta: MetaData{
			Title: "About",
			Description: 
				"The handful of ways by which the author defines himself.",
		},
	}
	ad.Raw = template.HTML(blackfriday.Run(md))

	writeTemplate(
		[]string{
			"_templates/about.gohtml",
			"_templates/base.gohtml",
		}, "about.html", ad)
}

func genRss(id *IndexData) {
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}

	feed := &feeds.Feed{
		Title: "tacix.at",
        Link: &feeds.Link{Href: "http://tacix.at"},
        Description: "Personal projects in programming and the dark arts..",
        Author: &feeds.Author{Name: "Taci Xat"},
        Created: time.Date(2020, time.May, 18, 12, 18, 0, 0, location),
        Updated: time.Now(),
	}

	for _, p := range id.Posts {
		link := fmt.Sprintf("http://tacix.at/%s.html", p.Meta.Title)

		feed.Items = append(feed.Items, &feeds.Item{
			Title: p.Meta.Title,
			Link: &feeds.Link{Href: link},
			Description: p.Meta.Description,
			Author: &feeds.Author{Name: "tacixat"},
			Created: p.Date,
			Content: string(p.Raw),
		})
	}

	f, err := os.Create("rss.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
	}

	f.Write([]byte(rss))
}

func main() {
	id := &IndexData{
		Meta: MetaData{
			Title:       "Home",
			Description: "A very rad blog.",
		},
		Posts: []PostData{},
	}

	genPosts(id)
	genRss(id)

	writeTemplate(
		[]string{
			"_templates/index.gohtml",
			"_templates/base.gohtml",
		}, "index.html", id)

	genAbout()
}
