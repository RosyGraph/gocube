package main

import "regexp"

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
