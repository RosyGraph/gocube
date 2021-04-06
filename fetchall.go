// Fetchall fetches URLs in parallel and reports their times and sizes.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type Cards struct {
	CardNames []Card `json:"cards"`
}

type Legalities struct {
	format   string
	legality string
}

type Card struct {
	Name     string  `json:"name"`
	CMC      float64 `json:"cmc"`
	ManaCost string  `json:manaCost`
	Text     string  `json:text`
}

// type Card struct {
// name         string
// manaCost     string
// cmc          int
// cardType     string `json:"type"`
// types        []string
// rarity       string
// set          string
// setName      string
// text         string
// artist       string
// number       int
// layout       string
// multiverseid string
// imageUrl     string
// printings    []string
// originalText string
// originalType string
// legalities   []*Legalities
// id           string
// }

func main() {
	start := time.Now()
	cardnames := []string{
		"Pyramids",
		"Farmstead",
		"Tundra",
		"Gitaxian Probe",
	}
	ch := make(chan string)
	for _, cardname := range cardnames {
		query := url.QueryEscape(cardname)
		url := fmt.Sprintf("https://api.magicthegathering.io/v1/cards?name=%q", query)
		go fetch(url, ch) // start a goroutine
	}
	for range cardnames {
		fmt.Println(<-ch) // receive from channel
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func fetch(url string, ch chan<- string) {
	start := time.Now()
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

	var s Cards
	err = json.Unmarshal(buffer, &s)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

	elapsed := time.Since(start).Seconds()
	card := s.CardNames[0]
	colors := colorID(card)
	ch <- fmt.Sprintf("%.2fs\t%s:\n\t\tcmc: %v\n\t\tmanacost: %s\n\tcolor identity:%v", elapsed, card.Name, card.CMC, card.ManaCost, colors)
}

func colorID(c Card) []string {
	colors := make([]string, 5)
	p := `\{([A-Z]/)*W([A-Z]/)*\}`
	if regexp.MatchString(p, c.ManaCost) || regex.MatchString(p, c.Text) {
		colors := colors.append("W")
	}
	return colors
}
