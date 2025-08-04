package extension

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"path/filepath"
	"regexp"
	"strings"
)

// -------- AST Transformer --------

type TransformerImage struct{}

func (t *TransformerImage) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if img, ok := n.(*ast.Image); ok {
				var origClass string
				var origTitle string
				var origSrc string
				for _, attr := range img.Attributes() {
					if string(attr.Name) == "class" {
						if v, ok := attr.Value.([]byte); ok {
							origClass = string(v)
						}
					} else if string(attr.Name) == "title" {
						if v, ok := attr.Value.([]byte); ok {
							origTitle = string(v)
						}
					} else if string(attr.Name) == "src" {
						if v, ok := attr.Value.([]byte); ok {
							origSrc = string(v)
						}
					}
				}
				if !strings.Contains(origClass, "img-fluid") {
					newClass := strings.TrimSpace(origClass + " img-fluid")
					img.SetAttributeString("class", []byte(newClass))
				}
				if origTitle == "" {
					img.SetAttributeString("title", filepath.Base(origSrc))
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

// -------- Renderer --------

type RendererImage struct{ html.Config }

var (
	imgTagRE      = regexp.MustCompile(`(?i)<img\b([^>]*)>`)
	imgTagClassRE = regexp.MustCompile(`(?i)class\s*=\s*"([^"]*)"`)
)

func NewImageRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &RendererImage{Config: html.NewConfig()}
	for _, o := range opts {
		o.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *RendererImage) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// covers both inline and block raw HTML
	reg.Register(ast.KindRawHTML, r.renderImgTag)
}

func (r *RendererImage) renderImgTag(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		raw := node.(*ast.RawHTML)
		orig := raw.Segments.Value(source)

		modified := imgTagRE.ReplaceAllFunc(orig, func(tag []byte) []byte {
			if imgTagClassRE.Match(tag) {
				return imgTagClassRE.ReplaceAllFunc(tag, func(m []byte) []byte {
					parts := imgTagClassRE.FindSubmatch(m)
					exist := ""
					if len(parts) > 1 {
						exist = string(parts[1])
					}
					if !strings.Contains(exist, "img-fluid") {
						exist = strings.TrimSpace(exist + " img-fluid")
					}
					return []byte(fmt.Sprintf(`class="%s"`, exist))
				})
			}
			s := string(tag)
			return []byte(strings.Replace(s, "<img", `<img class="img-fluid"`, 1))
		})

		_, _ = w.Write(modified)
	}
	return ast.WalkContinue, nil
}

// -------- Extension --------

type ImageExtension struct{}

func NewImageExtension() goldmark.Extender {
	return &ImageExtension{}
}

func (e *ImageExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(&TransformerImage{}, 999)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewImageRenderer(), 500)))
}
