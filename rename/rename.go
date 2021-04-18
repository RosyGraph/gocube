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
		path := path.Join("draftlogs", dir.Name())

		f, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file %s: %s\n", path, err.Error())
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		if err != nil {
			log.Printf("Error opening file %s: %s\n", path, err.Error())
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
		fmt.Printf("%s\n\t-> renamed to: %s", dir.Name(), newName)
	}
}
