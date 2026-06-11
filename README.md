## Overview
A super lightweight template engine to load and parse go templates from the local filesystem. It works with gin out of the box.

To install it:

```bash
go install github.com/gsuntres/pages
```

## Usage

Pages was designed to work with gin's HTMLRender.

```bash
engine := gin.New()

instance := pages.NewPagesWithProps(&pages.PagesProps{
  Mode: ModeLocal,
})

engine.HTMLRender = instance
```

By default, *Pages* will look for template files in the ```pages``` directory. To render a template you call

```go
c.HTML(200, "mypage.html")
```

or

```go
c.HTML(200, "subpath/mypage.html")
```

when paths relative to pages.

To render the page using a layout

```go
c.HTML(200, "subpath/mypage.html?layout=layout.html")
```

*Pages* will make sure to load layout.html first then mypage.html and render the final template.

## Cache

By default caching is disabled. Any changes on the templates will show up on the next render.

To enable it, pass WithCache = true to *PagesProps*.

## Helper Functions

Gotemplate's equivalent to locals in express.js is ```template.FuncMap```. *Pages* has the following functions availabe in templates:

* safe: to escape html
* safeURL: to escape url strings
* json: to parse data into json
* jsonPretty: same as json but more redable json
* ifelse: a convenient one line if else statement.

To add more

```go
instance.AddFunc(name, func)
```