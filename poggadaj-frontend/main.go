package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	_ "poggadaj-shared"
	"time"

	"poggadaj-shared/logging"

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
	i18n      map[string]*map[string]string
}

func newTemplateRenderer() (*TemplateRenderer, error) {
	// Load templates
	templates := map[string]*template.Template{}
	templateNames := []string{"html/home.html", "html/downloads.html"}
	for _, v := range templateNames {
		tmpl, err := template.New("").Funcs(template.FuncMap{
			"translate": func(m map[string]string, key string) string {
				// First try to get from selected mapping
				if val, ok := m[key]; ok {
					return val
				}
				// TODO: fallback to other language
				return key
			},
		}).ParseFiles("html/base.html", v)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s: %w", v, err)
		}

		templates[v] = tmpl
	}

	// Load i18n data
	i18n := map[string]*map[string]string{}
	files, err := filepath.Glob("i18n/*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	for _, v := range files {
		file, err := os.Open(v)
		if err != nil {
			// Ignore for now
			continue
		}

		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			continue
		}

		key := ""

		// TODO: Parse filenames here instead of hardcoding these
		switch v {
		case "i18n/en.json":
			key = "en"
		case "i18n/pl.json":
			key = "pl"
		}

		i18n[key] = &map[string]string{}

		err = json.Unmarshal(data, i18n[key])
		if err != nil {
			continue
		}
	}

	return &TemplateRenderer{templates: templates, i18n: i18n}, nil
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, passedData any) error {
	if tmpl, ok := t.templates[name]; ok {
		data := map[string]any{
			"i18n": t.i18n["en"],
			"data": passedData,
		}
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	return errors.New("couldn't find template")
}

func main() {
	logging.L = log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Level:           log.DebugLevel,
	})

	e := echo.New()
	e.Logger = slog.New(logging.L)

	var err error
	e.Renderer, err = newTemplateRenderer()
	if err != nil {
		panic(err)
	}

	e.Static("/static", "static")

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "html/home.html", nil)
	})

	e.GET("/download", func(c *echo.Context) error {
		clients := []HtmlClient{
			{
				Name:               "Gadu-Gadu 6.0",
				DescriptionI18nTag: "gg60-description",
				ImageUrl:           "../static/gg60.png",
				DownloadUrl:        "https://example.com",
			},
			{
				Name:               "Gadu-Gadu 7.7",
				DescriptionI18nTag: "gg77-description",
				ImageUrl:           "../static/gg77.png",
				DownloadUrl:        "https://example.com",
			},
		}
		return c.Render(http.StatusOK, "html/downloads.html", map[string]any{"Clients": clients})
	})

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("shutting down the server", "error", err)
	}
}
