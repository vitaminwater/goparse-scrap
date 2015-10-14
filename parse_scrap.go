package pscrap

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/vitaminwater/goparse"

	"github.com/gorilla/css/scanner"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
	"github.com/moovweb/gokogiri/xpath"
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
		o = parse.NewObject()
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

func (s *Scrapper) AddPage(client *http.Client, matcher Matcher) *Page {
	page := newPage(client, matcher)
	s.pages = append(s.pages, page)
	return page
}

type FieldPath []string

type pObjectField struct {
	path     FieldPath
	selector Selector
}

type Page struct {
	client     *http.Client
	matcher    Matcher
	fields     []*pObjectField
	processors []Processor
}

func newPage(client *http.Client, matcher Matcher) *Page {
	if client == nil {
		client = &http.Client{}
	}
	page := Page{client, matcher, []*pObjectField{}, []Processor{}}
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
		value, err := field.selector(url, doc)
		if err != nil {
			return err
		}
		o.SetNested(field.path, value)
	}

	for _, processor := range p.processors {
		processor(o)
	}
	return nil
}

func (p *Page) AddField(path FieldPath, selector Selector) {
	field := pObjectField{path, selector}
	p.fields = append(p.fields, &field)
}

func (p *Page) AddProcessor(processor Processor) {
	p.processors = append(p.processors, processor)
}

/**
 *	page matchers
 */

type Matcher func(string) bool

func RegexpMatcher(expr string) Matcher {
	return func(url string) bool {
		return false
	}
}

func HostMatcher(host string) Matcher {
	return func(rawUrl string) bool {
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

type Selector func(url string, doc *html.HtmlDocument) (interface{}, error)

type xpathSelectorApply func(match string, value interface{}) interface{}

func xpathSelector(xs []string, apply xpathSelectorApply) Selector {
	exprs := []*xpath.Expression{}
	for _, x := range xs {
		exprs = append(exprs, xpath.Compile(x))
	}
	return func(url string, doc *html.HtmlDocument) (interface{}, error) {
		var value interface{}
		for _, expr := range exprs {
			matches, err := doc.EvalXPath(expr, nil)
			if err != nil {
				return nil, err
			}

			if nodeset, ok := matches.([]xml.Node); ok == true {
				for _, node := range nodeset {
					value = apply(node.Content(), value)
				}
			} else {
				switch match := matches.(type) {
				case float64:
					value = apply(strconv.FormatFloat(match, 'f', 10, 64), value)
				case bool:
					if match {
						value = apply("true", value)
					} else {
						value = apply("false", value)
					}
				case string:
					value = apply(match, value)
				}
			}

		}

		return value, nil
	}
}

func XpathStringSelector(xs []string, sep string) Selector {
	selector := xpathSelector(xs, func(match string, value interface{}) interface{} {
		if value == nil {
			return match
		}
		vs := value.(string)
		vs = fmt.Sprintf("%s%s%s", vs, sep, match)
		return vs
	})
	return selector
}

func XpathStringArraySelector(xs []string) Selector {
	selector := xpathSelector(xs, func(match string, value interface{}) interface{} {
		if value == nil {
			return []interface{}{match}
		}
		vs := value.([]interface{})
		vs = append(vs, match)
		return vs
	})
	return selector
}

func XpathNumberSelector(xs string) Selector {
	selector := xpathSelector([]string{xs}, func(match string, value interface{}) interface{} {
		if n, err := strconv.ParseFloat(match, 64); err == nil {
			value = n
		} else {
			fmt.Println("Failed to parse Float in ", match, " reason ", err)
		}
		return value
	})
	return selector
}

/**
 *	Selector Middlewares. TODO: make chaining function
 */

func CssProperty(name string, selector Selector) Selector {
	return func(url string, doc *html.HtmlDocument) (interface{}, error) {
		value, err := selector(url, doc)
		if err != nil {
			return value, err
		}
		if v, ok := value.(string); ok == true {
			cssMap := CssToMap(v)
			if cssValue, ok := cssMap[name]; ok == true {
				return cssValue, nil
			}
			return nil, fmt.Errorf("Missing %s css property in %s", name, value)
		}
		return value, err
	}
}

func StripBlanks(selector Selector) Selector {
	return func(url string, doc *html.HtmlDocument) (interface{}, error) {
		value, err := selector(url, doc)
		if err != nil {
			return value, err
		}

		if v, ok := value.(string); ok == true {
			return strings.TrimSpace(v), nil
		} else if v, ok := value.([]interface{}); ok == true {
			for i, s := range v {
				if s, ok := s.(string); ok == true {
					v[i] = strings.TrimSpace(s)
				}
			}
			return v, nil
		}

		return value, err
	}
}

/**
 *	Processors
 */

type Processor func(o *parse.Object)

// Misc

func CssToMap(css string) map[string]string {
	res := map[string]string{}
	s := scanner.New(css)
	key := ""
	for {
		token := s.Next()
		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			break
		}
		if len(key) == 0 {
			if token.Type == scanner.TokenIdent {
				key = token.Value
			}
		} else if token.Type != scanner.TokenS && token.Type != scanner.TokenChar && token.Type != scanner.TokenComment {
			res[key] = token.Value
		} else if token.Type == scanner.TokenChar && token.Value == ";" {
			key = ""
		}
	}
	return res
}
