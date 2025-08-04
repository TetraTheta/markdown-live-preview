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
	"path/filepath"
	"regexp"
)

// -------- AST node --------

type HSImg struct {
	ast.BaseInline
	Src string
}

var KindHSImg = ast.NewNodeKind("HSImg")

func (n *HSImg) Kind() ast.NodeKind {
	return KindHSImg
}

func (n *HSImg) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Src": n.Src,
	}, nil)
}

// -------- Inline Parser --------

type ParserHSImg struct{}

var (
	imgSyntaxRE = regexp.MustCompile(`^\{\{<\s*img\s+"([^"]+?)"\s*>}}`)
	imgHTMLRE   = regexp.MustCompile(`\{\{<\s*img\s+"([^"]+?)"\s*>}}`)
)

func (p *ParserHSImg) Trigger() []byte {
	return []byte{'{'}
}

func (p *ParserHSImg) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	src := block.Source()
	pos, _ := block.Position()
	segment := src[pos:]
	m := imgSyntaxRE.FindSubmatch(segment)
	if m == nil {
		return nil
	}

	raw := string(m[1])
	if !regexp.MustCompile(`^[a-z]+://`).MatchString(raw) {
		if ext := path.Ext(raw); ext == "" {
			raw += ".webp"
		}
	}

	node := &HSImg{Src: raw}
	block.Advance(len(m[0]))
	return node
}

func (p *ParserHSImg) CloseBlock(ast.Node, text.Reader, parser.Context) {}

func (p *ParserHSImg) CanInterruptParagraph() bool {
	return false
}

func (p *ParserHSImg) CanAcceptIndentedLine() bool {
	return false
}

// -------- Renderer --------

type RendererHSImg struct{ html.Config }

func NewHSImgRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &RendererHSImg{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *RendererHSImg) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHSImg, r.renderHSImg)
}

func (r *RendererHSImg) renderHSImg(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*HSImg)

	esc := html2.EscapeString(n.Src)
	_, _ = w.WriteString(fmt.Sprintf(`<img src="%s" class="img-fluid" title="%s" />`, esc, filepath.Base(esc)))
	return ast.WalkContinue, nil
}

// -------- Post Processor --------

func PostProcessHSImg(input bytes.Buffer) string {
	replaced := imgHTMLRE.ReplaceAllFunc(input.Bytes(), func(match []byte) []byte {
		groups := imgHTMLRE.FindSubmatch(match)
		if len(groups) < 2 {
			return match
		}
		src := html2.EscapeString(string(groups[1]))
		return []byte(fmt.Sprintf(`<img src="%s" class="img-fluid" title="%s" />`, src, filepath.Base(src)))
	})
	return string(replaced)
}

// -------- Extension --------

type HSImgExtension struct{}

func NewHSImgExtension() goldmark.Extender {
	return &HSImgExtension{}
}

func (e *HSImgExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(util.Prioritized(&ParserHSImg{}, 500)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewHSImgRenderer(), 500)))
}
