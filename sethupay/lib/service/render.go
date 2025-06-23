package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

func newTemplateCache(templateRoot string) (map[string]*template.Template, error) {

	cache := map[string]*template.Template{}

	// Generate cache of web page templates
	pages, err := filepath.Glob(templateRoot + "/pages/*.go.html")
	if err != nil {
		return nil, fmt.Errorf("error generating list of templates in pages: %w", err)
	}

	tSet, err := template.ParseGlob(templateRoot + "/partials/*.go.html")
	if err != nil {
		return nil, fmt.Errorf("error generating partial templates for pages: %w", err)
	}
	tSet = tSet.Funcs(template.FuncMap{
		"withDashes": projectTemplate,
	})

	for _, page := range pages {

		name := filepath.Base(page)
		files := []string{
			templateRoot + "/common/base.go.html",
			templateRoot + "/common/head.go.html",
			templateRoot + "/common/top-menu.go.html",
			templateRoot + "/common/footer.go.html",
			templateRoot + "/common/js-includes.go.html",
			page,
		}
		tSet, err = tSet.ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("error creating template set for %s: %w", page, err)
		}

		cache[name] = tSet
	}

	// Append cache of email templates
	emails, err := filepath.Glob(templateRoot + "/emails/*.go.html")
	if err != nil {
		return nil, fmt.Errorf("error generating list of templates in emails: %w", err)
	}

	tEmailSet, err := template.ParseGlob(templateRoot + "/emails/partials/*.go.html")
	if err != nil {
		return nil, fmt.Errorf("error generating partial templates for emails: %w", err)
	}

	for _, email := range emails {

		name := filepath.Base(email)
		tEmailSet, err = tEmailSet.ParseFiles(email)
		if err != nil {
			return nil, fmt.Errorf("error creating template set for %s: %w", email, err)
		}

		cache[name] = tEmailSet
	}

	return cache, nil
}

func (s *Service) render(w http.ResponseWriter, template string, data any, status int) error {

	// Check whether that template exists in the cache
	tmpl, ok := s.Template[template]
	if !ok {
		return fmt.Errorf("template %s is not available in the cache", template)
	}

	var b bytes.Buffer
	if err := tmpl.ExecuteTemplate(&b, "base", data); err != nil {
		return fmt.Errorf("error executing template %s: %w", template, err)
	}

	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/html")
	w.Write(b.Bytes())

	return nil
}

// func (s *Service) renderEmail(template string, data any) (bytes.Buffer, error) {

// 	var emailBuf bytes.Buffer
// 	// Check whether that template exists in the cache
// 	tmpl, ok := s.Template[template]
// 	if !ok {
// 		return emailBuf, fmt.Errorf("template %s is not available in the cache", template)
// 	}

// 	if err := tmpl.ExecuteTemplate(&emailBuf, "email", data); err != nil {
// 		return emailBuf, fmt.Errorf("error executing template %s: %w", template, err)
// 	}

// 	return emailBuf, nil
// }

func (s *Service) renderJSON(w http.ResponseWriter, data []byte, status int) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	w.Write(data)

	return nil
}

func projectTemplate(in string) string {

	re := regexp.MustCompile(`\s+`)
	withDashes := re.ReplaceAllLiteralString(in, `-`)
	return strings.ToLower(withDashes)
}
