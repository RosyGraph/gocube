// Gocube processes and analyzes vintage cube draft picks.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Drafter struct {
	name string
	cmc  float64
	w    float64
	u    float64
	b    float64
	r    float64
	g    float64
}

var wg sync.WaitGroup

func main() {
	drafters := []string{"Jorbas", "RosyGraph", "Waluigi"}
	start := time.Now()
	for _, d := range drafters {
		analyzePicks(d)
	}
	fmt.Printf("total time:\t%.2f\n", time.Since(start).Seconds())
}

func analyzePicks(drafter string) {
	fmt.Printf("begin draft analysis for %s", drafter)
	drafts := processDraftPicks(drafter, "draftlogs")
	colors := map[string]float64{
		"W": 0.0,
		"U": 0.0,
		"B": 0.0,
		"R": 0.0,
		"G": 0.0,
		"X": 0.0,
	}
	records := [][]string{
		{"date", "W", "U", "B", "R", "G", "X"},
	}
	n := 0
	var cmc float64

	for i, draft := range drafts {
		fmt.Print(".")
		ch := make(chan Card, len(draft))

		for _, cardname := range draft {
			q := fmt.Sprintf("https://api.magicthegathering.io/v1/cards?name=%s", url.QueryEscape(cardname))
			go processCard(q, ch)
			wg.Add(1)
			n++
		}
		wg.Wait()
		close(ch)
		for c := range ch {
			cmc += c.CMC
			colorIDs := colorID(c)
			if len(colorIDs) == 0 {
				colors["X"]++
			}
			for _, color := range colorIDs {
				colors[color]++
			}
		}
		nf := float64(n)
		records = append(records, []string{
			strconv.FormatInt(int64(i), 10),
			strconv.FormatFloat(colors["W"]*100.0/nf, 'f', 2, 64),
			strconv.FormatFloat(colors["U"]*100.0/nf, 'f', 2, 64),
			strconv.FormatFloat(colors["B"]*100.0/nf, 'f', 2, 64),
			strconv.FormatFloat(colors["R"]*100.0/nf, 'f', 2, 64),
			strconv.FormatFloat(colors["G"]*100.0/nf, 'f', 2, 64),
			strconv.FormatFloat(colors["X"]*100.0/nf, 'f', 2, 64),
		})
	}

	fmt.Printf("done.\ndraft report for %s\n", drafter)
	fmt.Printf("avg cmc:\t%.2f\n", cmc/float64(n))
	fmt.Println("color preferences")
	for _, k := range []string{"W", "U", "B", "R", "G", "X"} {
		v := colors[k]
		fmt.Printf("%s:\t%.2f%%\n", k, v*100.0/float64(n))
	}

	f, err := os.Create(drafter + ".csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func processCard(s string, ch chan Card) {
	defer wg.Done()
	var resp *http.Response
	for i := 0; i < 5; i++ {
		if resp, err := http.Get(s); err != nil {
			fmt.Printf("%s: %s\n", s, err.Error())
		} else {
			if resp.StatusCode == 200 {
				buffer, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					continue
				}

				var cards Cards
				err = json.Unmarshal(buffer, &cards)
				if err != nil {
					continue
				}

				if len(cards.CardNames) == 0 {
					continue
				}
				ch <- cards.CardNames[0]
				resp.Close = true
				return
			}
		}
	}
	if resp != nil {
		resp.Close = true
	}
	fmt.Printf("[ERROR]\tgiving up on %s\n", s)
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
