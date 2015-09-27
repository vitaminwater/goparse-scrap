package pscrap

import (
	"fmt"
	"net/url"
	"net/http"
	"io/ioutil"

	"git.ccsas.biz/zilia_parse"

	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xpath"
	"github.com/moovweb/gokogiri/html"
)

type Scrapper struct {
	pages []*Page
}

func NewScrapper() *Scrapper {
	scrapper := Scrapper{[]*Page{}}

	return &scrapper
}

func (s *Scrapper) Scrap(url string, o *parse.Object) (*parse.Object, error) {
	page, err := s.pageForUrl(url)
	if err != nil {
		return nil, err
	}

	if o == nil {
		o = &parse.Object{}
	}

	if err := page.scrap(url, o); err != nil {
		return o, err
	}
	return o, nil
}

func (s *Scrapper) pageForUrl(url string) (*Page, error) {
	for _, page := range s.pages {
		if page.matcher(url) == true {
			return page, nil
		}
	}
	return nil, fmt.Errorf("Failed to find suitable page scrapper")
}

func (s *Scrapper) AddPage(p *Page) {
	s.pages = append(s.pages, p)
}

type FieldPath []string

type pObjectField struct {
	path FieldPath
	selector Selector
}

type Page struct {
	client *http.Client
	matcher Matcher
	fields []*pObjectField
}

func NewPage(client *http.Client, matcher Matcher) *Page {
	page := Page{client, matcher, []*pObjectField{}}
	return &page
}

func (p *Page) scrap(url string, o *parse.Object) error {
	resp, err := p.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	doc, err := gokogiri.ParseHtml(page)
	if err != nil {
		return err
	}
	defer doc.Free()

	for _, field := range p.fields {
		value, err := field.selector(doc)
		if err != nil {
			return err
		}
		o.SetNested(field.path, value)
	}
	return nil
}

func (p *Page) AddField(path FieldPath, selector Selector) {
	field := pObjectField{path, selector}
	p.fields = append(p.fields, &field)
}

/**
 *	page matchers
 */

type Matcher func(string) bool

func RegexpMatcher(expr string) Matcher {
	return func (url string) bool {
		return false
	}
}

func HostMatcher(host string) Matcher {
	return func (rawUrl string) bool {
		u, err := url.Parse(rawUrl)	
		if err != nil {
			fmt.Println(err)
			return false
		}
		return u.Host == host
	}
}

/**
 *	selectors
 */

type Selector func(doc *html.HtmlDocument) (interface{}, error)

func XpathSelector(xs []string, sep string) Selector {
	exprs := []*xpath.Expression{}
	for _, x := range xs {
		exprs = append(exprs, xpath.Compile(x))
	}
	return func (doc *html.HtmlDocument) (interface{}, error) {
		value := ""
		xdoc := xpath.NewXPath(doc.DocPtr())

		for _, expr := range exprs {
			err := xdoc.Evaluate(doc.Root().NodePtr(), expr)
			if err != nil {
				return nil, err
			}

			res, err := xdoc.ResultAsString()
			if err != nil {
				return nil, err
			}

			if len(value) == 0 {
				value = res
			} else {
				value = fmt.Sprintf("%s%s%s", value, sep, res)
			}
		}

		return value, nil
	}
}
