// Package pages simplifies work with templates. 
// It introduces convensions to better organize templates and streamline rendering.
//
// Pages expects the following directory structure:
//
// root/
// ├── items/
//     ├── [id].html
//     ├── index.html
// ├── index.html
// ├── layout.html
// ├── p1.html
// ├── p2.html
//
// pages.PageParse("root") will generate a root PageGroup and its nodes.

package pages

import (
	"os"
	"log"
	"slices"
	"strings"
	"net/url"
	"sync"
	"html/template"
	"path/filepath"
		
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	"git.gsuntres.com/gsuntres/pkg/sys"
)

var mu sync.RWMutex

var tplhtml map[string]*template.Template

type Mode int

const (
	ModeDefault = iota
	ModeLocal
	ModeS3
)

// Pages is the root instance
type Pages struct {
	Mode Mode

	WithCache bool

	// template files to be rendered on template load
	files map[string][]string

	templates map[string]*template.Template

	notFound *template.Template

	Root *PageGroup

	// Pages directory where templates are located (default: pages)
	Pages string
	
	funcMap template.FuncMap
}

func (p *Pages) AddFunc(name string, fn any) {
	mu.Lock()
	defer mu.Unlock()
	p.funcMap[name] = fn
}

func (p *Pages) GetFuncMap() template.FuncMap {
	mu.RLock()
	defer mu.RUnlock()
	
	return p.funcMap
}

// Count the number templates
func (p *Pages) Count() int {
	return len(p.templates)
}

func NewPages() *Pages {
	return NewPagesWithProps(nil)
}

type PagesProps struct {
	Mode Mode
	WithCache bool
	Pages string
}
 
func NewPagesWithProps(props *PagesProps) *Pages {
	o := &Pages{
		Mode: ModeDefault,
		files: map[string][]string{},
		funcMap: template.FuncMap{},
	}

	if props != nil && sys.FilePathValid(props.Pages) {
		o.Pages = props.Pages
	}

	if props == nil {
		props = &PagesProps{
			Mode: ModeDefault,
			WithCache: false,
		}
	}

	if err := o.Init(props); err != nil {
		log.Fatalf("Failed to init %v", err)
	}

	return o
}

func (p *Pages) BootstrapGin(ginEngin *gin.Engine, group *PageGroup) {
	Bootstrap(ginEngin, group)
}

func (p *Pages) AddTemplatesFromGroup(group *PageGroup) error {
	if group.IsRoot() {
		p.Root = group
	}

	var err error

	if group.HasIndex() {
    if err = p.AddTemplate(group.Index, group.GetLayout(), group.Index); err != nil {
      log.Printf("Failed to add template %v", err)

      return err
    }

    // add to files for template loading
    if _, indexFileOk := p.files[group.Index]; !indexFileOk {
    	p.files[group.Index] = make([]string, 0, 0)
    }
    if group.GetLayout() != "" {
    	p.files[group.Index] = append(p.files[group.Index], group.GetLayout())
    }
    p.files[group.Index] = append(p.files[group.Index], group.Index)
	}

  for _, page := range group.Pages {
    fullpath := page.AbsPath()
    if err = p.AddTemplate(fullpath, page.Layout(), fullpath); err != nil {
      log.Printf("Failed to add template %v", err)

      return err
    }

    if _, pageOk := p.files[fullpath]; !pageOk {
    	p.files[fullpath] = make([]string, 0, 0)
    }
    if page.Layout() != "" {
    	p.files[fullpath] = append(p.files[fullpath], page.Layout())
    }
    p.files[fullpath] = append(p.files[fullpath], fullpath)
	}

	for _, childGroup := range group.Groups {
		if err = p.AddTemplatesFromGroup(childGroup); err != nil {
			return nil
		}
  }

	return nil
}

func (p *Pages) Init(props *PagesProps) error {
	// mode
	p.Mode = props.Mode
	log.Printf("Mode %d", p.Mode)

	// cache
	p.WithCache = props.WithCache

	if p.WithCache {
		log.Printf("Template cache enabled")
	} else {
		log.Printf("Template cache disabled")
	}

	var err error
	p.notFound, err = template.New("not_found").Parse(`{{define "not_found"}}Page not found{{end}}`)
	if err != nil {
		return err
	}

	p.templates = make(map[string]*template.Template)

	p.AddFunc("safe", safeHTML)
	p.AddFunc("safeURL", safeHTML)
	p.AddFunc("json", jsonFunc)
	p.AddFunc("jsonPretty", jsonPretty)
	p.AddFunc("ifelse", ifelse)
	p.AddFunc("call", callGenerate(p))

	return nil
}

func (r *Pages) AddTemplate(name string, filenames... string) error {
	if r.templates == nil {
		r.templates = make(map[string]*template.Template)
	}

	root := template.New(name)

	usable := slices.DeleteFunc(filenames, func(n string) bool {
		return n == ""
	})

	// Load template files
	for _, filename := range usable {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", filename, err)

			continue
		}

		if _, err := root.Parse(string(content)); err != nil {
			log.Fatalf("Failed to parse %v", err)

			continue
		}
	}

	r.templates[name] = root

	return nil
}

func (p *Pages) Instance(name string, data any) render.Render {
	log.Printf("Render %s with %v", name, data)

	// render not_found
	if name == "not_found" {
		return render.HTML{
			Template: p.notFound,
			Name: "not_found",
			Data: gin.H{},
		}
	}
	
	// load template
	t := p.loadTemplate(name)

	if t == nil {
		return render.HTML{
			Template: p.notFound,
			Name: "not_found",
			Data: gin.H{},
		}
	}

	return render.HTML{
		Template: t,
		Name: name,
		Data: data,
	}
}

func (p *Pages) loadTemplate(path string) *template.Template {
	if p.WithCache {
		t, ok := p.templates[path]
		if ok {
			return t
		}
	}

	fullpath, err := url.Parse(path)
	if err != nil {
		log.Printf("Failed to parse url %v", path)

		return nil
	}

	q := fullpath.Query()

	root := template.New(path).Funcs(p.GetFuncMap())

	page := fullpath.Path
	layout := q.Get("layout")
	
	switch p.Mode {
	case ModeLocal:
		root = p.LoadLocal(root, layout, page)
	case ModeS3:
		log.Println("should load from s3", layout, page)
	default:
		root = p.LoadDefault(root, path)
	}

	if p.WithCache {
		p.templates[path] = root
	}

	return root
}

func (p *Pages) LoadDefault(root *template.Template, path string) *template.Template {
	log.Printf("Loading %s", path)

	files, ok := p.files[path]
	if !ok {
		return nil
	}

	log.Printf("Files %v", files)

	tpl := template.New(path)

	for _, fl := range files {
		data, err := os.ReadFile(fl)
		if err != nil {
			log.Fatalf("failed to load template: %v", err)
		}

		var errParse error
		tpl, errParse = tpl.Parse(string(data))
		if errParse != nil {
			log.Fatalf("failed to parse template: %v", errParse)
		}
	}

	return tpl
}

func (p *Pages) LoadLocal(root *template.Template, filenames... string) *template.Template {
	usable := slices.DeleteFunc(filenames, func(n string) bool {
		return strings.TrimSpace(n) == ""
	})

	for _, fl := range usable {
		fullPath := filepath.Join(p.Pages, fl)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatalf("failed to load template: %v", err)
		}

		var errParse error
		root, errParse = root.Parse(string(data))
		if errParse != nil {
			log.Fatalf("failed to parse template: %v", errParse)
		}
	}

	return root
}