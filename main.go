package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const homePage = "http://www.internetradiouk.com/"
const apiURL = "https://api.webrad.io/data/streams/42/"

var channels = map[string]string{}

func main() {
	getChanels(homePage)
	getAudioURL()
	for channel, url := range channels {
		log.Println(channel, url)
	}
}

func getChanels(home string) {
	resp, err := http.Get(home)
	if err != nil {
		panic(err)
	}
	doc := html.NewTokenizer(resp.Body)
	prefix := "http://www.internetradiouk.com/#"
	for tokenType := doc.Next(); tokenType != html.ErrorToken; {
		token := doc.Token()
		if tokenType == html.StartTagToken {
			if token.DataAtom != atom.A {
				tokenType = doc.Next()
				continue
			}

			for _, attr := range token.Attr {
				if attr.Key == "href" && strings.Contains(token.String(), prefix) {
					channels[attr.Val] = ""
					break
				}
			}
		}
		tokenType = doc.Next()
	}
	resp.Body.Close()
}

type radio struct {
	Station struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"station"`
	Streams []struct {
		ID          int    `json:"id"`
		IsContainer bool   `json:"isContainer"`
		MediaType   string `json:"mediaType"`
		Mime        string `json:"mime"`
		URL         string `json:"url"`
	} `json:"streams"`
}

func getAudioURL() {
	wg := sync.WaitGroup{}
	for c := range channels {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			rad := radio{}
			url := c[strings.LastIndex(c, "#")+1:]
			resp, err := http.Get(apiURL + url)
			if err != nil {
				panic(err)
			}
			err = json.NewDecoder(resp.Body).Decode(&rad)
			if err != nil {
				panic(err)
			}
			for _, st := range rad.Streams {
				if st.Mime == "audio/mpeg" {
					channels[c] = st.URL
				}
			}
		}(c)
	}
	wg.Wait()
}
