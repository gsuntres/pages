package pages

import (
	"html/template"
	"encoding/json"
	"log"
)

func ifelse (condition bool, a, b string) string {
	if condition {
		return a
	} else {
		return b
	}
}

func jsonFunc (o map[string]any) string {
	mJson, err := json.Marshal(o)
	if err != nil {
		log.Printf("unable to object %v", err)

		return ""
	}

	return string(mJson)
}

func jsonPretty (o map[string]any) string {
	mJson, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Printf("unable to object %v", err)

		return ""
	}

	return string(mJson)
}

func safeHTML(s string) template.HTML {
    return template.HTML(s)
}

func safeURL(s any) template.URL {
	if s == nil {
		return template.URL("")
	}

	v := s.(string)

	return template.URL(v)
}


func callGenerate(p *Pages) (func (name string, args ...any) any) {
	return func (name string, args ...any) any {
		fn, exists := p.funcMap[name]
		if !exists {
			return ""
		}

		if len(args) == 0 {
			if f, ok := fn.(func() string); ok { return f() }
		}

		if len(args) == 1 {
			if strArg, ok := args[0].(string); ok {
				if f, ok := fn.(func(string) string); ok { 
					return f(strArg) 
				}
			}
		}

		return "Invalid signature"
	}
}