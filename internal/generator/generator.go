// Package generator produces standardized Go source files for handlers, services, ports, adapters, models, and tests.
package generator

import (
	"errors"
	"strings"
	"unicode"
)

// ErrFileExists is returned when a generate target already exists on disk.
var ErrFileExists = errors.New("file already exists")

type genData struct {
	FileName        string
	Name            string
	Slug            string
	NameTitle       string
	NamePlural      string
	Module          string
	APIPackage      string
	HandlersPackage string
	ModelsPackage   string
	ServicesPackage string
	PortsPackage    string
	AdaptersPackage string
	APIImport       string
	HandlersImport  string
	ModelsImport    string
	ServicesImport  string
	PortsImport     string
	AdaptersImport  string
	ConfigImport    string
}

func splitNameTokens(s string) []string {
	var tokens []string
	var current []rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(string(current)))
		current = nil
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()

	if len(tokens) == 0 {
		return []string{"resource"}
	}
	return tokens
}

// newData builds the template data for a resource name. Module is filled by
// the caller (projectData uses the project's real module — no CWD reads).
func newData(name string) genData {
	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "_", "-"))
	tokens := strings.Split(slug, "-")
	plural := slug + "s"
	if strings.HasSuffix(slug, "s") {
		plural = slug
	}
	return genData{
		Name:       strings.ReplaceAll(slug, "-", "_"),
		Slug:       slug,
		FileName:   strings.ReplaceAll(slug, "-", "_"),
		NameTitle:  title(tokens),
		NamePlural: plural,
	}
}

// title capitalizes each token (replaces deprecated strings.Title).
func title(tokens []string) string {
	var b strings.Builder
	for _, tk := range tokens {
		if tk == "" {
			continue
		}
		b.WriteString(strings.ToUpper(tk[:1]) + tk[1:])
	}
	return b.String()
}
