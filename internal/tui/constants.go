package tui

import "github.com/sarkarshuvojit/commitlore/internal/core/llm"

// Content format constants - using constants from llm package
const (
	ContentFormatBlogArticle   = llm.ContentFormatBlogArticle
	ContentFormatTwitterThread = llm.ContentFormatTwitterThread
	ContentFormatLinkedInPost  = llm.ContentFormatLinkedInPost
	ContentFormatTechnicalDocs = llm.ContentFormatTechnicalDocs
)

// Content format descriptions
const (
	ContentFormatBlogArticleDesc   = "Long-form technical article suitable for dev.to, Medium, or personal blog"
	ContentFormatTwitterThreadDesc = "Engaging tweet series optimized for Twitter's format and audience"
	ContentFormatLinkedInPostDesc  = "Professional posts for LinkedIn networking and thought leadership"
	ContentFormatTechnicalDocsDesc = "Comprehensive technical documentation with architecture, APIs, and implementation details"
)