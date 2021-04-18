package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

func main() {
	dirs, _ := ioutil.ReadDir("draftlogs")
	regex, _ := regexp.Compile(`\d{4}-\d\d-\d\d_\d\d-\d\d`)
	for _, dir := range dirs {
		p := path.Join("draftlogs", dir.Name())

		f, err := os.Open(p)
		if err != nil {
			log.Printf("Error opening file %s: %s\n", p, err.Error())
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		if err != nil {
			log.Printf("Error opening file %s: %s\n", p, err.Error())
		}

		var drafter string
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "--> ") {
				drafter = strings.Replace(sc.Text(), "--> ", "", 1)
				break
			}
		}
		date := regex.FindString(dir.Name())
		newName := date + "_" + drafter + ".txt"
		os.Rename(p, path.Join("draftlogs", newName))
		fmt.Printf("%s\n\t-> renamed to: %s\n\n", dir.Name(), newName)
	}
}
