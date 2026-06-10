package pages

import (
	"html/template"
	"encoding/json"
	"log"
	"sync"
)
var mu sync.RWMutex
var funcMap = template.FuncMap{}

func AddFunc(name string, fn any) {
	mu.Lock()
	defer mu.Unlock()
	funcMap[name] = fn
}

func GetFuncMap() template.FuncMap {
	mu.RLock()
	defer mu.RUnlock()
	
	return funcMap
}

func init() {
	AddFunc("safe", safeHTML)
	AddFunc("safeURL", safeHTML)
	AddFunc("json", jsonFunc)
	AddFunc("jsonPretty", jsonPretty)
	AddFunc("ifelse", ifelse)
}

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
