Introduction
===

Scrap data directly to a [Parse](http://parse.com) backend.

It can also be used whenever you want to automate scrapping pages to
objects.

API
===

The main goal of this project is to be able to easily map datas found on
web pages to an object's fields.
Which in it's most simple form, is mostly assigning results of an
`XPath` expression to an object's field.

When you use `goparse-scrap` you first start by creating a Scrapper object
and add `Page` objects to it.
Then specify the fields that will be found on this particular page (or
type of pages).

A `Page` object is created by calling `AddPage` on the `Scrapper` with a `Matcher` as
parameter, a `Matcher` is actually just a function that returns true
when the current url matches the type of pages that the `Page` object
can handle.

For example:
```
scrapper := pscrap.NewScrapper()

// HostMatcher matches urls by host
// in this case, all urls with www.seloger.com as host will match
page := scrapper.AddPage(client, pscrap.HostMatcher("www.seloger.com"))

```

The function `pscrap.HostMatcher` actually returns the `Matcher`:

```

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

```

Now that we a setup our `Page` object, we want it to be able to set
values in our object.
Fields are added by calling the `AddField` method on a `Page` object, it
takes two arguments, a `FieldPath` which is the path to the object
field, you can set nested objects fields by specifying multiples keys;
the second argument to `AddField` is a `Selector`, again, just like
`Matcher`, a `Selector` is just a function that takes the `url` and the
fetched `HtmlDocument` as arguments and returns an `interface{}` value
that will be set to the field described by the `FieldPath` argument.


```

// Set the field test with the value found when executing the given XPath
page.AddField(pscrap.FieldPath{"test"}, pscrap.XpathStringArraySelector([]string{"//*[@id=\"slider1\"]/li/img/@src"}))

```

`XpathStringArraySelector` returns a function that executes the given
XPaths and sets the results as a `[]string` for the `test` key in the object.


Now that our `Page` is setup and has a field, we can start scrapping by
calling the `Scrap` method on our `Scrapper` object. This method takes
to arguments, a url, `testpage` in the following example, and an
`Object`, if the `Object` is nil, it will create one automatically.

```

// the host of this page is www.seloger.com so our page will catch it
testpage := "http://www.seloger.com/annonces/achat/appartement/paris-1er-75/101834495.htm?cp=75001&idtt=2&idtypebien=1&listing-listpg=4&tri=d_dt_crea&bd=Li_LienAnn_1"

if o, err := scrapper.Scrap(testpage, nil); err == nil {
  fmt.Println(o)
} else {
  panic(err)
}

```

In this example, the returned object `o` has a `test` field which is a `[]string`
of the results of the XPath found on the page.

The `Object` type is defined in the (goparse)[github.com/vitaminwater/goparse] project,
further documentation can be found there.

Middlewares
===

Because the goparse-scrap API uses functions, middlewares is a free
feature, you just have to chain them.
Chainging functions just means that you take one in parameter, and
return a new one that calls the first one.

For example, here is a `Selector` middleware, `CssProperty` that returns
a css property value. It is created by calling the `CssProperty` that
takes a `name` argument for the css property to retrieve, and a
`Selector` argument, that is the previous `Selector` in the chain;
and returns a `Selector` that can be used as-is with the `AddField`
method.
What this new `Selector` does is just call the given `Selector`, and
then tries to parse the result as a css expression to extract the given
css property, and returns this css property in place of the first
`Selector`'s value.

```

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

```

Post process
===

Sometimes, after the object has been scrapped, you still have some
operations to do. Maybe you want to be able to generate a net field
based on scrapped fields values.

You can do this by creating a `Processor` object, as usual, a
`Processor` is just a function the takes the currently scrapped Object
as parameter.
You add a `Processor` by calling the `AddProcessor` method on the `Page`
object, all processors are then called after each successful scrapings.
