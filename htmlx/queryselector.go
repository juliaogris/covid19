package htmlx

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	ErrInvalidSelector = fmt.Errorf("invalid selector")
	ErrDuplicateID     = fmt.Errorf("duplicate id")
	ErrInvalidTag      = fmt.Errorf("invalid tag")
)

func QuerySelector(n *html.Node, selector string) (*html.Node, error) {
	s, err := ParseSelectors(selector)
	if err != nil {
		return nil, err
	}
	if len(s) > 1 {
		return QueryNested(n, s), nil
	}
	return Query(n, s[0]), nil
}

func QueryNested(r *html.Node, selectors []*Selector) *html.Node {
	selectorCnt := len(selectors)
	if selectorCnt == 0 {
		return nil
	} else if selectorCnt == 1 {
		return Query(r, selectors[0])
	}
	nodes := QueryAll(r, selectors[0])
	for _, n := range nodes {
		if QueryNested(n, selectors[1:]) != nil {
			return n
		}
	}
	return nil
}

func Query(r *html.Node, s *Selector) *html.Node {
	if r == nil || s == nil {
		return nil
	}
	if Match(r, s) {
		return r
	}
	for c := r.FirstChild; c != nil; c = c.NextSibling {
		if n := Query(c, s); n != nil {
			return n
		}
	}
	return nil
}

func QueryAll(r *html.Node, s *Selector) []*html.Node {
	if r == nil || s == nil {
		return nil
	}
	return queryAll(r, s, true)
}

func QueryAllNoChildren(r *html.Node, s *Selector) []*html.Node {
	if r == nil || s == nil {
		return nil
	}
	return queryAll(r, s, false)
}

func queryAll(r *html.Node, s *Selector, queryChildrenOnMatch bool) []*html.Node {
	var result []*html.Node
	if Match(r, s) {
		if !queryChildrenOnMatch {
			return []*html.Node{r}
		}
		result = append(result, r)
	}
	for c := r.FirstChild; c != nil; c = c.NextSibling {
		if n := queryAll(c, s, queryChildrenOnMatch); n != nil {
			result = append(result, n...)
		}
	}
	return result
}

func Match(n *html.Node, s *Selector) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if s.Tag != "" && s.Tag != n.Data {
		return false
	}
	if s.ID != "" && !hasID(n, s.ID) {
		return false
	}
	return hasClasses(n, s.Classes)
}

func hasClasses(n *html.Node, classes []string) bool {
	if len(classes) == 0 {
		return true
	}
	for _, a := range n.Attr {
		if a.Key == "class" {
			classVal := a.Val
			for _, c := range classes {
				if !strings.Contains(classVal, c) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func hasID(n *html.Node, id string) bool {
	for _, a := range n.Attr {
		if a.Key == "id" {
			return a.Val == id
		}
	}
	return false
}

func ValidateSelector(s *Selector) error {
	if s.Tag != "" && !validTag[s.Tag] {
		return fmt.Errorf("%w: %s", ErrInvalidTag, s.Tag)
	}
	if s.Tag == "" && s.ID == "" && len(s.Classes) == 0 {
		return fmt.Errorf("%w: nothing specified: %#v", ErrInvalidSelector, s)
	}
	return nil
}

type Selector struct {
	Tag     string
	ID      string
	Classes []string
}

const token = "[.#]?-?[_a-zA-Z]+[_a-zA-Z0-9-]*"

var tokenRe = regexp.MustCompile(token)
var selectorRe = regexp.MustCompile("^(" + token + ")+$")

func ParseSelectors(s string) ([]*Selector, error) {
	selectors := strings.Fields(s)
	if len(selectors) == 0 {
		return nil, fmt.Errorf("no selector %s", s)
	}
	result := make([]*Selector, len(selectors))
	var err error
	for i, s := range selectors {
		result[i], err = ParseSelector(s)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ParseSelector(s string) (*Selector, error) {
	if !selectorRe.MatchString(s) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSelector, s)
	}
	sl := &Selector{}
	for _, e := range tokenRe.FindAllString(s, -1) {
		c := e[0]
		switch {
		case c == '.':
			sl.Classes = append(sl.Classes, e[1:])
		case c == '#' && sl.ID == "":
			sl.ID = e[1:]
		case c == '#' && sl.ID != "":
			return nil, fmt.Errorf("%w: '%s'", ErrDuplicateID, e[1:])
		case sl.Tag == "":
			sl.Tag = strings.ToLower(e)
		default:
			return nil, fmt.Errorf("%w: bad token %s", ErrInvalidSelector, e)
		}
	}
	if err := ValidateSelector(sl); err != nil {
		return nil, err
	}
	return sl, nil
}

// Copied from https://github.com/sindresorhus/html-tags
// MIT license
var validTag = map[string]bool{
	"a":          true,
	"abbr":       true,
	"address":    true,
	"area":       true,
	"article":    true,
	"aside":      true,
	"audio":      true,
	"b":          true,
	"base":       true,
	"bdi":        true,
	"bdo":        true,
	"blockquote": true,
	"body":       true,
	"br":         true,
	"button":     true,
	"canvas":     true,
	"caption":    true,
	"cite":       true,
	"code":       true,
	"col":        true,
	"colgroup":   true,
	"data":       true,
	"datalist":   true,
	"dd":         true,
	"del":        true,
	"details":    true,
	"dfn":        true,
	"dialog":     true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"em":         true,
	"embed":      true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"form":       true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"head":       true,
	"header":     true,
	"hgroup":     true,
	"hr":         true,
	"html":       true,
	"i":          true,
	"iframe":     true,
	"img":        true,
	"input":      true,
	"ins":        true,
	"kbd":        true,
	"label":      true,
	"legend":     true,
	"li":         true,
	"link":       true,
	"main":       true,
	"map":        true,
	"mark":       true,
	"math":       true,
	"menu":       true,
	"menuitem":   true,
	"meta":       true,
	"meter":      true,
	"nav":        true,
	"noscript":   true,
	"object":     true,
	"ol":         true,
	"optgroup":   true,
	"option":     true,
	"output":     true,
	"p":          true,
	"param":      true,
	"picture":    true,
	"pre":        true,
	"progress":   true,
	"q":          true,
	"rb":         true,
	"rp":         true,
	"rt":         true,
	"rtc":        true,
	"ruby":       true,
	"s":          true,
	"samp":       true,
	"script":     true,
	"section":    true,
	"select":     true,
	"slot":       true,
	"small":      true,
	"source":     true,
	"span":       true,
	"strong":     true,
	"style":      true,
	"sub":        true,
	"summary":    true,
	"sup":        true,
	"svg":        true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"template":   true,
	"textarea":   true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"time":       true,
	"title":      true,
	"tr":         true,
	"track":      true,
	"u":          true,
	"ul":         true,
	"var":        true,
	"video":      true,
	"wbr":        true,
}
