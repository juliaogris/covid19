package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func scrape(url string) (*data, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer closeIgnoreErr(resp.Body)
	return processHTML(resp.Body)
}

func processHTML(r io.Reader) (*data, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	tbody, err := findTbody(doc)
	if err != nil {
		return nil, err
	}

	return parseTrs(tbody)
}

func findTbody(n *html.Node) (*html.Node, error) {
	if isNode(n, "div", "table_container") {
		return unwrapTbody(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		n, err := findTbody(c)
		if err != nil {
			return nil, err
		}
		if n != nil {
			return n, nil
		}
	}
	return nil, nil
}

func unwrapTbody(div *html.Node) (*html.Node, error) {
	div2 := firstChild(div, "div")
	table := firstChild(div2, "table")
	if table == nil {
		return nil, fmt.Errorf(`unwrapTbody: expected <table>, got nil in %#v`, div2)
	}
	thead := firstChild(table, "thead")
	if err := validateThead(thead); err != nil {
		return nil, err
	}
	return next(thead, "tbody"), nil
}

func validateThead(thead *html.Node) error {
	if !isNode(thead, "thead", "") {
		return fmt.Errorf(`validateThead: expected <thead>, got %#v`, thead)
	}
	tr := firstChild(thead, "tr")
	if tr == nil {
		return fmt.Errorf(`validateThead: expected <tr>, got nil in %#v`, thead)
	}
	thTexts := []string{"Location", "Confirmed cases", "Cases per 1M people", "Recovered", "Deaths"}
	th := firstChild(tr, "th")

	for _, tt := range thTexts {
		if !isNode(th, "th", "") {
			return fmt.Errorf(`validateThead: expected <th>, got %#v`, tr)
		}
		tn := th.FirstChild
		if tn == nil || tn.Type != html.TextNode {
			return fmt.Errorf(`validateThead: expected <th> child to be TextNode, got %#v`, tn)
		}
		if tn.Data != tt {
			return fmt.Errorf(`validateThead: expected TextNode %s, got %s`, tt, tn.Data)
		}
		th = next(th, "th")
	}
	return nil
}

func parseTrs(tbody *html.Node) (*data, error) {
	if tbody == nil {
		return nil, fmt.Errorf("parseTrs: tbody is nil")
	}
	d := data{
		date:    time.Now(),
		entries: []entry{},
	}
	for tr := firstChild(tbody, "tr"); tr != nil; tr = next(tr, "tr") {
		entry, err := parseTr(tr)
		if err != nil {
			return nil, err
		}
		d.entries = append(d.entries, *entry)
	}
	return &d, nil
}

func parseTr(tr *html.Node) (*entry, error) {
	e := entry{}
	var err error
	countryTd := firstChild(tr, "td")
	e.country, err = getCountry(countryTd)
	if err != nil {
		return nil, err
	}
	var cells []string
	for td := next(countryTd, "td"); td != nil; td = next(td, "td") {
		text, err := getText(td)
		if err != nil {
			return nil, err
		}
		text = strings.Trim(text, " \n\t")
		text = strings.ReplaceAll(text, ",", "") // strip "," in numbers
		if text == "-" || text == "—" {
			text = "0"
		}
		cells = append(cells, text)
	}
	if len(cells) < 4 {
		return nil, fmt.Errorf("parseTr: table row has %d cells, expected >=4", len(cells)+1)
	}
	e.cases, err = strconv.Atoi(cells[0])
	if err != nil {
		return nil, fmt.Errorf("parseTr: cannot parse cases integer %w, %s", err, cells[0])
	}
	e.casesPer1M, err = strconv.ParseFloat(cells[1], 64)
	if err != nil {
		return nil, fmt.Errorf("parseTr: cannot parse casesPer1M float %w, %s", err, cells[1])
	}
	e.deaths, err = strconv.Atoi(cells[2])
	if err != nil {
		return nil, fmt.Errorf("parseTr: cannot parse deaths integer %w, '%s'", err, cells[2])
	}
	e.recoveries, err = strconv.Atoi(cells[3])
	if err != nil {
		return nil, fmt.Errorf("parseTr: cannot parse recoveries integer %w, %s", err, cells[3])
	}
	return &e, nil
}

func getCountry(td *html.Node) (string, error) {
	if !isNode(td, "td", "") {
		return "", fmt.Errorf(`getCountry: expected <td>, got %#v`, td)
	}
	span := firstChild(td, "span")
	if !isNode(span, "span", "") {
		return "", fmt.Errorf(`getCountry: expected <span>, got %#v`, span)
	}
	if span.FirstChild == nil || span.FirstChild.Type != html.TextNode {
		return "", fmt.Errorf(`getCountry: expected 1 text child for <span>, got %#v`, span.FirstChild)
	}
	return span.FirstChild.Data, nil
}

func getText(td *html.Node) (string, error) {
	if !isNode(td, "td", "") {
		return "", fmt.Errorf(`getText: expected <td>, got %#v`, td)
	}
	if td.FirstChild == nil || td.FirstChild.Type != html.TextNode {
		return "", fmt.Errorf(`getText: expected 1 text child for <td>, got %#v`, td.FirstChild)
	}
	return td.FirstChild.Data, nil
}

func isNode(n *html.Node, tag, class string) bool {
	if n == nil {
		return false
	}
	if n.Type != html.ElementNode {
		return false
	}
	if n.Data != tag {
		return false
	}
	return hasClassAttr(n.Attr, class)
}

func firstChild(n *html.Node, tag string) *html.Node {
	if n == nil || n.FirstChild == nil {
		return nil
	}
	c := n.FirstChild
	if c.Type == html.ElementNode && c.Data == tag {
		return c
	}
	return next(c, tag)
}

func next(n *html.Node, tag string) *html.Node {
	if n == nil {
		return nil
	}

	for n := n.NextSibling; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == tag {
			return n
		}
	}
	return nil
}

func hasClassAttr(attr []html.Attribute, class string) bool {
	if class == "" {
		return true
	}
	for _, a := range attr {
		if a.Key == "class" {
			return strings.Contains(a.Val, class)
		}
	}
	return false
}

func closeIgnoreErr(c io.Closer) {
	_ = c.Close()
}
