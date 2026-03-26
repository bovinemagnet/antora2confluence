package renderer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformToConfluence_Paragraph(t *testing.T) {
	html := `<div class="paragraph"><p>Hello world.</p></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<p>Hello world.</p>")
}

func TestTransformToConfluence_Headings(t *testing.T) {
	html := `<h2 id="_section">Section Title</h2>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<h2>Section Title</h2>")
}

func TestTransformToConfluence_CodeBlock(t *testing.T) {
	html := `<div class="listingblock"><div class="content"><pre class="highlight"><code class="language-bash" data-lang="bash">echo hello</code></pre></div></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="code"`)
	assert.Contains(t, result, "echo hello")
	assert.Contains(t, result, "bash")
}

func TestTransformToConfluence_CodeBlock_NoLanguage(t *testing.T) {
	html := `<div class="listingblock"><div class="content"><pre class="highlight"><code>plain code</code></pre></div></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="code"`)
	assert.Contains(t, result, "plain code")
}

func TestTransformToConfluence_UnorderedList(t *testing.T) {
	html := `<div class="ulist"><ul><li><p>Item one</p></li><li><p>Item two</p></li></ul></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<ul>")
	assert.Contains(t, result, "<li>")
	assert.Contains(t, result, "Item one")
}

func TestTransformToConfluence_OrderedList(t *testing.T) {
	html := `<div class="olist arabic"><ol class="arabic"><li><p>First</p></li><li><p>Second</p></li></ol></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<ol>")
	assert.Contains(t, result, "First")
}

func TestTransformToConfluence_Table(t *testing.T) {
	html := `<table class="tableblock frame-all grid-all stretch"><colgroup><col style="width: 50%;"><col style="width: 50%;"></colgroup><thead><tr><th class="tableblock halign-left valign-top">Code</th><th class="tableblock halign-left valign-top">Meaning</th></tr></thead><tbody><tr><td class="tableblock halign-left valign-top"><p class="tableblock">400</p></td><td class="tableblock halign-left valign-top"><p class="tableblock">Bad Request</p></td></tr></tbody></table>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<table>")
	assert.Contains(t, result, "<th>")
	assert.Contains(t, result, "Code")
	assert.Contains(t, result, "400")
}

func TestTransformToConfluence_InlineFormatting(t *testing.T) {
	html := `<div class="paragraph"><p>This has <strong>bold</strong> and <em>italic</em> and <code>monospace</code> text.</p></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, "<strong>bold</strong>")
	assert.Contains(t, result, "<em>italic</em>")
	assert.Contains(t, result, "<code>monospace</code>")
}

func TestTransformToConfluence_Link(t *testing.T) {
	html := `<div class="paragraph"><p>Visit <a href="https://example.com">Example</a>.</p></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `href="https://example.com"`)
	assert.Contains(t, result, "Example")
}

func TestTransformToConfluence_Image(t *testing.T) {
	html := `<div class="imageblock"><div class="content"><img src="logo.png" alt="Logo"></div></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="image"`)
	assert.Contains(t, result, `ri:filename="logo.png"`)
}

func TestTransformToConfluence_InlineImage(t *testing.T) {
	html := `<div class="paragraph"><p>See <span class="image"><img src="icon.png" alt="Icon"></span> here.</p></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="image"`)
	assert.Contains(t, result, `ri:filename="icon.png"`)
}

func TestTransformToConfluence_AdmonitionNote(t *testing.T) {
	html := `<div class="admonitionblock note"><table><tr><td class="icon"><div class="title">Note</div></td><td class="content">This is a note.</td></tr></table></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="info"`)
	assert.Contains(t, result, "This is a note.")
}

func TestTransformToConfluence_AdmonitionWarning(t *testing.T) {
	html := `<div class="admonitionblock warning"><table><tr><td class="icon"><div class="title">Warning</div></td><td class="content">Be careful.</td></tr></table></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="warning"`)
	assert.Contains(t, result, "Be careful.")
}

func TestTransformToConfluence_AdmonitionTip(t *testing.T) {
	html := `<div class="admonitionblock tip"><table><tr><td class="icon"><div class="title">Tip</div></td><td class="content">A helpful tip.</td></tr></table></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `ac:name="tip"`)
	assert.Contains(t, result, "A helpful tip.")
}

func TestTransformToConfluence_Blockquote(t *testing.T) {
	html := `<div class="quoteblock"><blockquote><div class="paragraph"><p>A famous quote.</p></div></blockquote></div>`
	result, err := TransformToConfluence(html)
	require.NoError(t, err)
	assert.Contains(t, result, `<blockquote>`)
	assert.Contains(t, result, "A famous quote.")
}

func TestTransformToConfluence_EmptyInput(t *testing.T) {
	result, err := TransformToConfluence("")
	require.NoError(t, err)
	assert.Equal(t, "", result)
}
