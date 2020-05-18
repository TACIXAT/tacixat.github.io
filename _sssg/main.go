package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
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

type IndexData struct {
	Posts []string
}

type PostData struct {
	Title string
	Tags  []string
	Date  time.Time
	Raw   template.HTML
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

func main() {
	id := IndexData{
		Posts: []string{},
	}

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

		pd.Raw = template.HTML(blackfriday.Run(md))
		writeTemplate(
			[]string{
				"_templates/post.gohtml",
				"_templates/base.gohtml",
			}, fmt.Sprintf("posts/%s.html", pd.Title), pd)
		id.Posts = append(id.Posts, pd.Title)
	}

	writeTemplate(
		[]string{
			"_templates/index.gohtml",
			"_templates/base.gohtml",
		}, "index.html", id)
}
