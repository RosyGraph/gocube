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
	"sync"
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
var wg sync.WaitGroup

func main() {
	start := time.Now()
	cardnames := []string{
		"Plains",
		"Island",
		"Swamp",
		"Mountain",
		"Forest",
	}
	ch := make(chan Card, len(cardnames))

	for _, cardname := range cardnames {
		q := fmt.Sprintf("https://api.magicthegathering.io/v1/cards?name=%q", url.QueryEscape(cardname))
		go func(s string) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				if resp, err := http.Get(s); err != nil {
					fmt.Printf("%s: %s\n", s, err.Error())
					return
				} else {
					if resp.StatusCode == 200 {
						buffer, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, resp.Status)
							continue
						}

						var cards Cards
						err = json.Unmarshal(buffer, &cards)
						if err != nil {
							fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", cards.CardNames[0].Name, i, resp.Status)
							continue
						}

						if len(cards.CardNames) == 0 {
							fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", cards.CardNames[0].Name, i, cardname)
							continue
						}
						ch <- cards.CardNames[0]
						return
					} else {
						fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, resp.Status)
					}
				}
			}
			fmt.Printf("[ERROR]\tgiving up on %s\n", s)
		}(q)
		wg.Add(1)
	}
	wg.Wait()
	close(ch)
	fmt.Printf("total time: %.2f\n", time.Since(start).Seconds())

	var cmc float64
	for v := range ch {
		fmt.Println(v.Name)
		cmc += v.CMC
	}
	fmt.Printf("total cmc: %.2f\n", cmc)
}

func fetch(url string, ch chan<- Card) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err.Error())
		ch <- Card{Name: url, Text: err.Error()}
		return
	}
	defer resp.Body.Close()

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- Card{Name: url, Text: err.Error()}
		return
	}

	var s Cards
	err = json.Unmarshal(buffer, &s)
	if err != nil {
		ch <- Card{Name: url, Text: err.Error()}
		return
	}

	if len(s.CardNames) == 0 {
		ch <- Card{Name: url, Text: fmt.Sprintf("could not find %s", url)}
		return
	}

	card := s.CardNames[0]
	card.ColorID = colorID(card)
	ch <- card
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

		if err != nil {
			panic(err)
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "--> ") && !strings.HasSuffix(sc.Text(), drafter) {
				picks = append(picks, sc.Text()[4:])
			}
		}
	}
	return picks
}
