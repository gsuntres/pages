package pages

import (
	"log"
	"path"
	"path/filepath"
	"regexp"

	"git.gsuntres.com/gsuntres/pkg/sys"
)

// Page holds information required to render the final page and route
type Page struct {
	Group     *PageGroup
	Name      string
	Path      string
	Filename  string
}

func (page *Page) AbsPath() string {
	fullpath, _ := filepath.Abs(filepath.Join(page.Group.Dir, page.Filename))
	
	return fullpath
}

func (page *Page) FullPath() string {
	return filepath.Join(page.Group.FullPath(), page.Filename)
}

func (page *Page) HasLayout() bool {
	return page.Group.HasLayout()
}

func (page *Page) Layout() string {
	return ParentLayout(page.Group)
}

func (page *Page) GetPath() string {
	re := regexp.MustCompile(`\[(\w+)\]`)
	result := re.ReplaceAllString(page.Path, `:$1`)

	return result
}

func (page *Page) HasParams() bool {
	return len(page.GetParams()) > 0
}

func (page *Page) GetParams() []string {
	re := regexp.MustCompile(`\[(\w+)\]`)
	match := re.FindStringSubmatch(page.Path)

	return match
}

func ParentLayout(group *PageGroup) string {
	if group == nil {
		return ""
	}

	if group.Layout != "" {
		return group.Layout
	} else {
		return ParentLayout(group.Parent)
	}
}

// PageGroup provides a general view of the directory where templates and other directories exist.
type PageGroup struct {
	Dir    string
	Index  string
	Layout string
	Path   string
	Pages  []*Page
	Parent *PageGroup
	Groups []*PageGroup
}

func (pg *PageGroup) FullPath() string {
	var v []string

	for pg != nil {
		v = append(v, pg.Path)
		pg = pg.Parent
	}

	return filepath.Join(v...)
}

func (pg *PageGroup) HasLayout() bool {
	return pg.Layout != ""
}

func (pg *PageGroup) GetLayout() string {
	if pg.Layout != "" {
		return pg.Layout
	}

	if pg.Parent != nil {
		return pg.Parent.GetLayout()
	}

	return "__layout_not_found__"
}

func (pg *PageGroup) HasIndex() bool {
	return pg.Index != ""
}

func (pg *PageGroup) IsRoot() bool {
	return pg.Parent == nil
}

func PageParse(dir string) (*PageGroup, error) {
	return PageParseWithParent(dir, nil)
}

func PageParseWithParent(dir string, parent *PageGroup) (*PageGroup, error) {
	group := &PageGroup{
		Dir: dir,
		Parent: parent,
	}

	if parent == nil {
		group.Path = "/"
	} else {
		group.Path = path.Base(dir)
	}

	// Check index, layout
	indexHtml := filepath.Join(dir, "index.html")
	if sys.FileExists(indexHtml) {
		group.Index = indexHtml
	}

	layoutHtml := filepath.Join(dir, "layout.html")
	if sys.FileExists(layoutHtml) {
		group.Layout = layoutHtml
	}

	// Check other pages
	otherPages, err := sys.GetHtmlFiles(dir, "index.html", "layout.html")
	
	if err == nil {
		if len(otherPages) > 0 {
			group.Pages = make([]*Page, 0)
		}

		for _, opage := range otherPages {
			name := sys.FileBase(opage)
			
			page := &Page{
				Group: group,
				Name: name,
				Path: name,
				Filename: opage,
			}

			group.Pages = append(group.Pages, page)
		}
	}

	// Check directories
	otherDirs, err := sys.GetDirs(dir)
	if err != nil {
		return nil, err
	}

	if len(otherDirs) > 0 {
		group.Groups = make([]*PageGroup, 0)
	}

	for _, odir := range otherDirs {
		fdir := filepath.Join(dir, odir)
		childGroup, err := PageParseWithParent(fdir, group)
		if err != nil {
			log.Printf("failed to parse group :%v", err)

			continue
		}

		group.Groups = append(group.Groups, childGroup)
	}

	return group, nil
}
