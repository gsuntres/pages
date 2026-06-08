// Package pages simplifies work with templates. 
// It introduces convensions to better organize templates and streamline rendering.
//
// Pages expects the following directory structure:
//
// root/
// ├── items/
// 		 ├── [id].html
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
	"net/url"
	"html/template"
		
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

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
}

// func (p *Pages) FindPage(path string) *Page {
// 	return findRecursive(p.Root, path)
// }

// func findRecursive(group *PageGroup, path string) *Page {
// 	var found *Page
// 	for _, p := range group.Pages {
// 		if p.Path == path {
// 			found = p

// 			break
// 		}
// 	}

// 	if found != nil {
// 		return found
// 	}

// 	for _, gp := range group.Groups {
// 		found = findRecursive(gp, path)

// 		if found != nil {
// 			break
// 		}
// 	}

// 	return found
// }

func NewPages() *Pages {
	return NewPagesWithPros(&PagesProps{
		WithCache: false,
	})
}
 
func NewPagesWithPros(props *PagesProps) *Pages {
	o := &Pages{
		Mode: ModeDefault,
		files: map[string][]string{},
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

type PagesProps struct {
	WithCache bool
}

func (r *Pages) Init(props *PagesProps) error {
	r.WithCache = props.WithCache

	if r.WithCache {
		log.Printf("Template cache enabled")
	} else {
		log.Printf("Template cache disabled")
	}

	var err error
	r.notFound, err = template.New("not_found").Parse(`{{define "not_found"}}Page not found{{end}}`)
	if err != nil {
		return err
	}

	r.templates = make(map[string]*template.Template)

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
			log.Fatalf("Failed to parse template %v", err)

			continue
		}
	}

	r.templates[name] = root

	return nil
}

func (r *Pages) Instance(name string, data any) render.Render {
	log.Printf("Render %s with %v", name, data)
	
	// render not_found
	if name == "not_found" {
		return render.HTML{
			Template: r.notFound,
			Name: "not_found",
			Data: gin.H{},
		}
	}

	// load template
	t := r.loadTemplate(name)

	if t == nil {
		return render.HTML{
			Template: r.notFound,
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

	root := template.New(path).Funcs(funcMap)

	page := fullpath.Path
	layout := q.Get("layout")
	
	switch p.Mode {
	case ModeLocal:
		log.Println("should load from local", layout, page)
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
	log.Printf("Load %s", path)

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

	// // TODO temp
	// toimplement, _ := template.New("to_implement").Parse("Need to implement")
	// return toimplement
}