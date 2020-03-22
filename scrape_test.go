package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestIsNode(t *testing.T) {
	frag := `<div>hello</div>`
	r := strings.NewReader(frag)
	nodes, err := html.ParseFragment(r, nil)

	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	n := nodes[0].FirstChild.NextSibling.FirstChild // html.head.body.div
	require.True(t, isNode(n, "div", ""))
}

func TestIsNodeWithClass(t *testing.T) {
	frag := `<span class="big blue">hello</span>`
	r := strings.NewReader(frag)
	nodes, err := html.ParseFragment(r, nil)

	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	n := nodes[0].FirstChild.NextSibling.FirstChild // html.head.body.div
	require.True(t, isNode(n, "span", "blue"))
}

func TestProcessHTML(t *testing.T) {
	fpath := filepath.Join("testdata", "coronavirus-map-2020-03-22.html")
	r, err := os.Open(fpath)
	require.NoError(t, err)

	data, err := processHTML(r)
	require.NoError(t, err)

	require.Equal(t, 168, len(data.entries))
	want := entry{
		country:    "Worldwide",
		cases:      303594,
		casesPer1M: 43.09,
		deaths:     94625,
		recoveries: 12964,
	}
	require.Equal(t, want, data.entries[0])
}

func TestDataString(t *testing.T) {
	date, err := time.Parse("2006-01-02", "2020-03-22")
	require.NoError(t, err)
	d := data{
		date: date,
		entries: []entry{
			{"Austria", 2992, 336.08, 9, 8},
			{"Australia", 1122, 44.00, 43, 7},
		},
	}
	want := `
                         Austria     2992    336.1        9        8
                       Australia     1122     44.0       43        7`
	want = want[1:]
	require.Equal(t, want, d.String())
}

func TestScrapErr(t *testing.T) {
	_, err := scrape("not a URL")
	require.Error(t, err)
}

func TestProcessHTMLErr(t *testing.T) {
	r := strings.NewReader(`<div class="table_container"><div></div>no tabls</div>`)
	_, err := processHTML(r)
	require.Error(t, err)
}

func TestUnwrapTbody(t *testing.T) {
	r := strings.NewReader(`<html><head></head><body><div><table></table><div></div></div><body></html>`)
	n, err := html.Parse(r)
	require.NoError(t, err)
	body := n.FirstChild.FirstChild.NextSibling
	_, err = unwrapTbody(body)
	require.Error(t, err)
}

func TestValidateTheadErr(t *testing.T) {
	r := strings.NewReader(`<html><head></head><body><table><thead></thead><tbody></tobdy></table><div></div></div><body></html>`)
	n, err := html.Parse(r)
	require.NoError(t, err)
	body := n.FirstChild.FirstChild.NextSibling

	err = validateThead(body)
	require.Error(t, err)

	thead := body.FirstChild.FirstChild
	err = validateThead(thead)
	require.Error(t, err)
}

func TestValidateTheadThErr(t *testing.T) {
	format := `<html><head></head><body><table><thead><tr>%s</tr></thead><tbody></tobdy></table><div></div></div><body></html>`
	ss := []string{"", "<th><div></div></th>", "<th>WrongLabel</th>"}

	for _, s := range ss {
		r := strings.NewReader(fmt.Sprintf(format, s))
		n, err := html.Parse(r)
		require.NoError(t, err)
		thead := n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild
		err = validateThead(thead)
		require.Error(t, err)
	}
}

func TestNext(t *testing.T) {
	require.Nil(t, next(nil, "p"))
}

func TestParseTrsErr(t *testing.T) {
	_, err := parseTrs(nil)
	require.Error(t, err)

	r := strings.NewReader(`<html><head></head><body><table><tbody><tr></tr></tobdy></table><div></div></div><body></html>`)
	n, err := html.Parse(r)
	require.NoError(t, err)
	tbody := n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild
	_, err = parseTrs(tbody)
	require.Error(t, err)
}

func TestParseTrErr(t *testing.T) {
	_, err := parseTrs(nil)
	require.Error(t, err)

	format := `<html><head></head><body><table><tbody><tr><td><span>USA</span></td>%s</tr></tobdy></table><div></div></div><body></html>`
	ss := []string{
		"<td><div></div></td>",
		"<td>123</td>",
		"<td>notANumber</td><td>1.2</td><td>1</td><td>2</td>",
		"<td>1</td><td>not a number</td><td>1</td><td>2</td>",
		"<td>1</td><td>1.2</td><td>not a number</td><td>2</td>",
		"<td>1</td><td>1.2</td><td>1</td><td>not a number</td>",
	}
	for _, s := range ss {
		r := strings.NewReader(fmt.Sprintf(format, s))
		n, err := html.Parse(r)
		require.NoError(t, err)
		tr := n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild
		_, err = parseTr(tr)
		require.Error(t, err)
	}
}

func TestGetTextErr(t *testing.T) {
	_, err := getText(nil)
	require.Error(t, err)
}

func TestGetCountryErr(t *testing.T) {
	format := `<html><head></head><body><table><tbody><tr><td>%s</td></tr></tobdy></table><div></div></div><body></html>`
	ss := []string{"", "<span><div></div></span>"}
	for _, s := range ss {
		r := strings.NewReader(fmt.Sprintf(format, s))
		n, err := html.Parse(r)
		require.NoError(t, err)
		td := n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild.FirstChild

		_, err = getCountry(td)
		require.Error(t, err)
	}
}
