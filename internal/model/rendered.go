package model

// RenderedPage holds the output of rendering an AsciiDoc page
// into Confluence storage format.
type RenderedPage struct {
	// SourcePage is the original page that was rendered.
	SourcePage Page

	// Title extracted from the rendered output.
	Title string

	// Body is the Confluence storage format markup.
	Body string

	// Fingerprint is a SHA-256 hash of the normalised body + title,
	// used for change detection.
	Fingerprint string

	// Includes lists the resolved include file paths found during rendering.
	Includes []string

	// Images lists the image filenames referenced in the page.
	Images []string

	// XRefs lists the internal xref targets found in the page.
	XRefs []string
}
