package pscrap

import (
	"git.ccsas.biz/zilia_parse"
)

type Scrapper struct {
	pages []*Page
}

func NewScrapper() *Scrapper {
	scrapper := Scrapper{[]*Page{}}

	return &scrapper
}

func (s *Scrapper) Scrap(url string, o *parse.Object) (*parse.Object, error) {
	for _, page := range s.pages {
		if o, err := page.scrap(url, o); err != nil {
			return o, err
		}
	}
	return o, nil
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
	matcher Matcher
	fields []*pObjectField
}

func NewPage(matcher Matcher) *Page {
	page := Page{matcher, []*pObjectField{}}
	return &page
}

func (p *Page) scrap(url string, o *parse.Object) (*parse.Object, error) {
	return o, nil
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

/**
 *	selectors
 */

type Selector func() interface{}

func XpathSelector(xpaths []string, sep string) Selector {
	return func () interface{} {
		return nil
	}
}
