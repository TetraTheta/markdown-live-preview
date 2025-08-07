package extension

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	html2 "html"
	"path"
	"regexp"
	"strings"
)

// -------- AST node --------

type HSGalleryImage struct {
	ast.BaseBlock
	Src     []string
	Width   string
	Caption string
}

var KindHSGalleryImage = ast.NewNodeKind("HSGalleryImage")

func (g *HSGalleryImage) Kind() ast.NodeKind {
	return KindHSGalleryImage
}

func (g *HSGalleryImage) Dump(source []byte, level int) {
	ast.DumpHelper(g, source, level, map[string]string{
		"Src":     fmt.Sprintf("%v", g.Src),
		"Width":   g.Width,
		"Caption": g.Caption,
	}, nil)
}

// -------- Block Parser --------

type ParserHSGalleryImage struct{}

var (
	galleryImageSyntaxRE = regexp.MustCompile(`^\{\{<\s*gallery/image\s+([^>]+?)\s*>}}\s*$`)
	galleryImageAttrRE   = regexp.MustCompile(`(\w+)\s*=\s*"([^"]*)"`)
)

func (p *ParserHSGalleryImage) Trigger() []byte {
	return []byte{'{'}
}

func (p *ParserHSGalleryImage) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	m := galleryImageSyntaxRE.FindSubmatch(line)
	if m == nil {
		return nil, parser.NoChildren
	}

	attrsRaw := string(m[1])
	attrs := make(map[string]string)
	for _, sm := range galleryImageAttrRE.FindAllStringSubmatch(attrsRaw, -1) {
		attrs[sm[1]] = sm[2]
	}

	if _, hasWidth := attrs["width"]; !hasWidth {
		if v, ok := attrs["w"]; ok {
			attrs["width"] = v
		}
	}
	if _, hasCaption := attrs["caption"]; !hasCaption {
		if v, ok := attrs["c"]; ok {
			attrs["caption"] = v
		}
	}

	srcRaw, ok := attrs["src"]
	if !ok || strings.TrimSpace(srcRaw) == "" {
		return nil, parser.NoChildren
	}
	srcItems := strings.Split(srcRaw, "|")
	if len(srcItems) < 1 || len(srcItems) > 3 {
		return nil, parser.NoChildren
	}

	for i, srcItem := range srcItems {
		if !regexp.MustCompile(`^[a-z]+://`).MatchString(srcItem) {
			if ext := path.Ext(srcItem); ext == "" {
				srcItem += ".webp"
			}
		}
		srcItems[i] = strings.TrimSpace(srcItem)
	}

	node := &HSGalleryImage{
		Src:     srcItems,
		Width:   attrs["width"],
		Caption: attrs["caption"],
	}
	reader.Advance(len(line))
	return node, parser.HasChildren
}

func (p *ParserHSGalleryImage) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *ParserHSGalleryImage) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *ParserHSGalleryImage) CanInterruptParagraph() bool {
	return true
}

func (p *ParserHSGalleryImage) CanAcceptIndentedLine() bool {
	return false
}

// -------- Renderer --------

type RendererHSGalleryImage struct{ html.Config }

func NewHSGalleryImageRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &RendererHSGalleryImage{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *RendererHSGalleryImage) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHSGalleryImage, r.renderHSGalleryImage)
}

func (r *RendererHSGalleryImage) renderHSGalleryImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*HSGalleryImage)

	pct := "100%"
	if len(n.Src) == 2 {
		pct = "49.5%"
	} else if len(n.Src) == 3 {
		pct = "32.5%"
	}

	widthAttr := "100%"
	if n.Width != "" {
		widthAttr = html2.EscapeString(n.Width)
	}
	_, _ = w.WriteString(fmt.Sprintf(`<div class="gallery" align="center"><div style="width:%s">`, widthAttr))
	_, _ = w.WriteString(`<figure class="gallery-figure"><div class="gallery-container">`)

	for _, src := range n.Src {
		esc := html2.EscapeString(strings.TrimSpace(src))
		_, _ = w.WriteString(fmt.Sprintf(`<div class="gallery-inner" style="width:%s"><img src="%s" class="img-fluid" title="%s" /></div>`, pct, esc, html2.EscapeString(src)))
	}
	_, _ = w.WriteString(`</div>`)

	if n.Caption != "" {
		var buf = bytes.NewBuffer(nil)
		_ = goldmark.Convert([]byte(n.Caption), buf)
		_, _ = w.WriteString(fmt.Sprintf(`<figcaption class="gallery-figcaption">%s</figcaption>`, buf.String()))
	}

	_, _ = w.WriteString(`</figure></div></div>`)
	return ast.WalkContinue, nil
}

// -------- Extension --------

type HSGalleryImageExtension struct{}

func NewHSGalleryImageExtension() goldmark.Extender {
	return &HSGalleryImageExtension{}
}

func (e *HSGalleryImageExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(util.Prioritized(&ParserHSGalleryImage{}, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewHSGalleryImageRenderer(), 100)))
}
