// Fetchall fetches URLs in parallel and reports their times and sizes.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Cards struct {
	CardNames []Card `json:"cards"`
}

type Card struct {
	Name     string  `json:"name"`
	CMC      float64 `json:"cmc"`
	ManaCost string  `json:manaCost`
	Text     string  `json:text`
	ColorID  []string
}

type Drafter struct {
	name string
	cmc  float64
	w    float64
	u    float64
	b    float64
	r    float64
	g    float64
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
	cardnames := processDraftPicks("RosyGraph")
	picks := make([]Card, 0)

	ch := make(chan Card)
	for _, cardname := range cardnames {
		query := url.QueryEscape(cardname)
		url := fmt.Sprintf("https://api.magicthegathering.io/v1/cards?name=%q", query)
		go fetch(url, ch) // start a goroutine
	}
	for range cardnames {
		picks = append(picks, <-ch)
	}
	fmt.Printf("%v\n", picks)
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func fetch(url string, ch chan<- Card) {
	resp, err := http.Get(url)
	if err != nil {
		ch <- Card{Name: url, Text: err.Error()}
		return
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- Card{Name: url, Text: err.Error()}
		resp.Body.Close()
		return
	}

	var s Cards
	err = json.Unmarshal(buffer, &s)
	if err != nil {
		ch <- Card{Name: url, Text: err.Error()}
		resp.Body.Close()
		return
	}

	if len(s.CardNames) == 0 {
		ch <- Card{Name: url, Text: err.Error()}
		resp.Body.Close()
		return
	}

	card := s.CardNames[0]
	card.ColorID = colorID(card)
	ch <- card
	resp.Body.Close()
}

func colorID(c Card) []string {
	colors := make([]string, 0)
	patterns := map[string]string{
		"W": `\{([A-Z]\/)?W(\/[A-Z])?\}`,
		"U": `\{([A-Z]\/)?U(\/[A-Z])?\}`,
		"B": `\{([A-Z]\/)?B(\/[A-Z])?\}`,
		"R": `\{([A-Z]\/)?R(\/[A-Z])?\}`,
		"G": `\{([A-Z]\/)?G(\/[A-Z])?\}`,
	}

	for k, p := range patterns {
		if match, _ := regexp.MatchString(p, c.ManaCost+c.Text); match {
			colors = append(colors, k)
		}
	}

	return colors
}

func processDraftPicks(drafter string) []string {
	logs, err := ioutil.ReadDir("draftlogs")
	if err != nil {
		panic(err)
	}
	picks := make([]string, 0)

	for _, log := range logs {
		if !strings.Contains(log.Name(), drafter) {
			continue
		}
		f, err := os.Open("draftlogs/" + log.Name())
		defer f.Close()

		if err != nil {
			panic(err)
		}

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "--> ") && !strings.HasSuffix(sc.Text(), drafter) {
				picks = append(picks, sc.Text()[4:])
			}
		}
	}
	return picks
}
