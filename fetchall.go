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
	/*
	 * cardnames := []string{
	 *     "Plains",
	 *     "Island",
	 *     "Swamp",
	 *     "Mountain",
	 *     "Forest",
	 *     "Lightning Bolt",
	 *     "Giant Growth",
	 *     "Healing Salve",
	 *     "Dark Ritual",
	 *     "Ancestral Recall",
	 * }
	 */
	drafts := processDraftPicks("RosyGraph", "draftlogs")
	colors := map[string]float64{
		"W": 0.0,
		"U": 0.0,
		"B": 0.0,
		"R": 0.0,
		"G": 0.0,
		"X": 0.0,
	}
	n := 0
	var cmc float64

	for _, draft := range drafts {
		ch := make(chan Card, len(draft))

		for _, cardname := range draft {
			q := fmt.Sprintf("https://api.magicthegathering.io/v1/cards?name=%s", url.QueryEscape(cardname))
			go func(s string) {
				defer wg.Done()
				var resp *http.Response
				for i := 0; i < 5; i++ {
					if resp, err := http.Get(s); err != nil {
						fmt.Printf("%s: %s\n", s, err.Error())
					} else {
						if resp.StatusCode == 200 {
							buffer, err := ioutil.ReadAll(resp.Body)
							if err != nil {
								// fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, resp.Status)
								continue
							}

							var cards Cards
							err = json.Unmarshal(buffer, &cards)
							if err != nil {
								// fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, resp.Status)
								continue
							}

							if len(cards.CardNames) == 0 {
								// fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, cardname)
								continue
							}
							ch <- cards.CardNames[0]
							resp.Close = true
							return
						} else {
							// fmt.Printf("[ERROR]\t%s (tried %d times):\n\t%s\n", s, i, resp.Status)
						}
					}
				}
				resp.Close = true
				fmt.Printf("[ERROR]\tgiving up on %s\n", s)
			}(q)
			wg.Add(1)
			n++
		}
		wg.Wait()
		close(ch)
		fmt.Printf("total time:\t%.2f\n", time.Since(start).Seconds())
		for c := range ch {
			cmc += c.CMC
			for _, color := range colorID(c) {
				colors[color]++
			}
		}
	}

	fmt.Printf("avg cmc:\t%.2f\n", cmc/float64(n))
	fmt.Println("color preferences")
	for _, k := range []string{"W", "U", "B", "R", "G", "X"} {
		v := colors[k]
		fmt.Printf("%s:\t%.2f\n", k, v/float64(n))
	}
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

func processDraftPicks(drafter, dir string) [][]string {
	logs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	drafts := make([][]string, 0)

	for _, log := range logs {
		picks := make([]string, 0)
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
		drafts = append(drafts, picks)
	}
	return drafts
}
