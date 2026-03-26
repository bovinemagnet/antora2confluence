package renderer

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// PageTitleMap maps relative page paths (without extension) to their titles.
// Used to resolve internal xref links to Confluence page links.
type PageTitleMap map[string]string

// TransformToConfluence converts asciidoctor HTML5 output into
// Confluence storage format markup.
func TransformToConfluence(htmlContent string, pageTitles PageTitleMap) (string, error) {
	if htmlContent == "" {
		return "", nil
	}

	nodes, err := html.ParseFragment(
		strings.NewReader(htmlContent),
		&html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body},
	)
	if err != nil {
		return "", fmt.Errorf("parsing HTML: %w", err)
	}

	var buf bytes.Buffer
	for _, n := range nodes {
		transformNode(&buf, n, pageTitles)
	}

	return strings.TrimSpace(buf.String()), nil
}

func transformNode(buf *bytes.Buffer, n *html.Node, pageTitles PageTitleMap) {
	switch n.Type {
	case html.TextNode:
		buf.WriteString(n.Data)
		return
	case html.ElementNode:
		// handled below
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
		return
	}

	if n.Data == "div" {
		class := getAttr(n, "class")
		switch {
		case strings.Contains(class, "admonitionblock"):
			transformAdmonition(buf, n, class)
			return
		case strings.Contains(class, "listingblock"):
			transformCodeBlock(buf, n)
			return
		case strings.Contains(class, "imageblock"):
			transformImageBlock(buf, n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
		return
	}

	if n.Data == "img" {
		transformImage(buf, n)
		return
	}

	if n.Data == "span" && strings.Contains(getAttr(n, "class"), "image") {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
		return
	}

	if len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
		buf.WriteString("<" + n.Data + ">")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
		buf.WriteString("</" + n.Data + ">")
		return
	}

	if n.Data == "table" {
		transformTable(buf, n)
		return
	}

	switch n.Data {
	case "p", "ul", "ol", "li", "strong", "em", "code", "blockquote", "pre", "sup", "sub":
		buf.WriteString("<" + n.Data + ">")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
		buf.WriteString("</" + n.Data + ">")
	case "a":
		href := getAttr(n, "href")
		if pageTitles != nil && isInternalLink(href) {
			// Convert internal xref to Confluence page link
			baseName := strings.TrimSuffix(href, ".html")
			if title, ok := pageTitles[baseName]; ok {
				buf.WriteString(`<ac:link><ri:page ri:content-title="`)
				buf.WriteString(html.EscapeString(title))
				buf.WriteString(`"/>`)
				// Include link text as link body
				var linkText bytes.Buffer
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					extractText(&linkText, c)
				}
				if linkText.Len() > 0 {
					buf.WriteString(`<ac:link-body>`)
					buf.WriteString(linkText.String())
					buf.WriteString(`</ac:link-body>`)
				}
				buf.WriteString(`</ac:link>`)
			} else {
				// Internal link but title not found — render as plain link
				writePassthroughLink(buf, n, href, pageTitles)
			}
		} else {
			writePassthroughLink(buf, n, href, pageTitles)
		}
	case "br":
		buf.WriteString("<br/>")
	case "hr":
		buf.WriteString("<hr/>")
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			transformNode(buf, c, pageTitles)
		}
	}
}

func isInternalLink(href string) bool {
	if href == "" {
		return false
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "#") {
		return false
	}
	return strings.HasSuffix(href, ".html")
}

func writePassthroughLink(buf *bytes.Buffer, n *html.Node, href string, pageTitles PageTitleMap) {
	buf.WriteString(`<a`)
	if href != "" {
		buf.WriteString(fmt.Sprintf(` href="%s"`, href))
	}
	buf.WriteString(">")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		transformNode(buf, c, pageTitles)
	}
	buf.WriteString("</a>")
}

func transformAdmonition(buf *bytes.Buffer, n *html.Node, class string) {
	macroName := "info"
	switch {
	case strings.Contains(class, "warning"):
		macroName = "warning"
	case strings.Contains(class, "caution"):
		macroName = "warning"
	case strings.Contains(class, "tip"):
		macroName = "tip"
	case strings.Contains(class, "important"):
		macroName = "note"
	case strings.Contains(class, "note"):
		macroName = "info"
	}

	content := extractAdmonitionContent(n)

	buf.WriteString(fmt.Sprintf(`<ac:structured-macro ac:name="%s">`, macroName))
	buf.WriteString(`<ac:rich-text-body>`)
	buf.WriteString(fmt.Sprintf(`<p>%s</p>`, content))
	buf.WriteString(`</ac:rich-text-body>`)
	buf.WriteString(`</ac:structured-macro>`)
}

func extractAdmonitionContent(n *html.Node) string {
	td := findNodeByClass(n, "content")
	if td == nil {
		return ""
	}
	var buf bytes.Buffer
	for c := td.FirstChild; c != nil; c = c.NextSibling {
		extractText(&buf, c)
	}
	return strings.TrimSpace(buf.String())
}

func transformCodeBlock(buf *bytes.Buffer, n *html.Node) {
	codeNode := findNode(n, "code")
	if codeNode == nil {
		codeNode = findNode(n, "pre")
	}
	if codeNode == nil {
		return
	}

	lang := getAttr(codeNode, "data-lang")
	var codeBuf bytes.Buffer
	extractText(&codeBuf, codeNode)

	buf.WriteString(`<ac:structured-macro ac:name="code">`)
	if lang != "" {
		buf.WriteString(fmt.Sprintf(`<ac:parameter ac:name="language">%s</ac:parameter>`, lang))
	}
	buf.WriteString(`<ac:plain-text-body><![CDATA[`)
	buf.WriteString(codeBuf.String())
	buf.WriteString(`]]></ac:plain-text-body>`)
	buf.WriteString(`</ac:structured-macro>`)
}

func transformImageBlock(buf *bytes.Buffer, n *html.Node) {
	img := findNode(n, "img")
	if img == nil {
		return
	}
	transformImage(buf, img)
}

func transformImage(buf *bytes.Buffer, img *html.Node) {
	src := getAttr(img, "src")
	if src == "" {
		return
	}

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		buf.WriteString(fmt.Sprintf(`<ac:structured-macro ac:name="image"><ac:parameter ac:name="url">%s</ac:parameter></ac:structured-macro>`, src))
		return
	}

	filename := src
	if idx := strings.LastIndex(filename, "/"); idx >= 0 {
		filename = filename[idx+1:]
	}

	buf.WriteString(`<ac:structured-macro ac:name="image">`)
	buf.WriteString(fmt.Sprintf(`<ri:attachment ri:filename="%s"/>`, filename))
	buf.WriteString(`</ac:structured-macro>`)
}

func transformTable(buf *bytes.Buffer, n *html.Node) {
	buf.WriteString("<table>")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		switch c.Data {
		case "thead":
			buf.WriteString("<thead>")
			transformTableRows(buf, c, "th")
			buf.WriteString("</thead>")
		case "tbody":
			buf.WriteString("<tbody>")
			transformTableRows(buf, c, "td")
			buf.WriteString("</tbody>")
		}
	}
	buf.WriteString("</table>")
}

func transformTableRows(buf *bytes.Buffer, section *html.Node, cellTag string) {
	for tr := section.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.Data != "tr" {
			continue
		}
		buf.WriteString("<tr>")
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if td.Type != html.ElementNode {
				continue
			}
			tag := cellTag
			if td.Data == "th" {
				tag = "th"
			}
			buf.WriteString("<" + tag + ">")
			for c := td.FirstChild; c != nil; c = c.NextSibling {
				extractText(buf, c)
			}
			buf.WriteString("</" + tag + ">")
		}
		buf.WriteString("</tr>")
	}
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func findNode(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNode(c, tag); result != nil {
			return result
		}
	}
	return nil
}

func findNodeByClass(n *html.Node, class string) *html.Node {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == "class" && strings.Contains(a.Val, class) {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNodeByClass(c, class); result != nil {
			return result
		}
	}
	return nil
}

func extractText(buf *bytes.Buffer, n *html.Node) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(buf, c)
	}
}
