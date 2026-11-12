package filepolicy

import "path/filepath"

// ContentTypeForPath maps the assignment's accepted extensions to media types.
// An empty result means the extension is unsupported.
func ContentTypeForPath(path string) string {
	switch filepath.Ext(path) {
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}
