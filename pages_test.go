package pages

import (
	"os"
	"path/filepath"
	"bytes"
	"testing"

	"github.com/gin-gonic/gin/render"

	"git.gsuntres.com/gsuntres/pkg/commons"
)

func TestNew_Nil(t *testing.T) {
	pages := NewPagesWithProps(nil)

	if pages == nil {
		t.Fatal("Should have created pages")
	}
}

func TestNew_DefaultProps(t *testing.T) {
	pages := NewPagesWithProps(nil)

	if pages.WithCache != false {
		t.Fatal("Default without cache")
	}

	if pages.Mode != ModeDefault {
		t.Fatal("Mode should have been default")
	}
}

func TestNew_InitTemplates(t *testing.T) {
	pages := NewPagesWithProps(nil)

	if pages.Count() != 0 {
		t.Fatal("Should had no templates")
	}
}

func TestAddTemplate(t *testing.T) {
	pages := NewPagesWithProps(nil)
	sampleTpl := filepath.Join(".test", "group1_id_rendering.html")
	pages.AddTemplate("group1_id_rendering", sampleTpl)

	if pages.Count() != 1 {
		t.Fatal("Should had one template")
	}
}

func TestInstance(t *testing.T) {
	pages := NewPagesWithProps(&PagesProps{
		Mode: ModeLocal,
		Pages: ".test/root",
	})
	
	ren := pages.Instance("group1/index.html?layout=layout.html", map[string]any{"foo": "bar"})
	h, ok := ren.(render.HTML)
	if !ok {
		t.Fatal("Failed to convert to render.HTML")
	}

	var out bytes.Buffer
	err := h.Template.Execute(&out, h.Data)
	if err != nil {
		t.Fatalf("Failed to execute template %v", err)
	}

	body := out.String()

	expectedBytes, _ := os.ReadFile("./.test/group1_rendering_index.html")
	expectedBody := string(expectedBytes)

	expected := commons.StringNormalize(expectedBody)
	actual := commons.StringNormalize(body)

	if actual != expected {
		t.Error("Unexpected rendered html")
	}
}