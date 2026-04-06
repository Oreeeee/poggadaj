package main

import (
	"context"
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
	"github.com/labstack/echo/v5/middleware"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
	i18n      map[string]*map[string]string
}

func newTemplateRenderer() (*TemplateRenderer, error) {
	// Load templates
	templates := map[string]*template.Template{}
	templateNames := []string{"html/home.html", "html/downloads.html", "html/login.html"}
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
			"safeValueAccess": func(m map[string]any, key string) any {
				if val, ok := m[key]; ok {
					return val
				}

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
			"lang": "en",
			"data": passedData,
		}
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	return errors.New("couldn't find template")
}

func main() {
	var err error
	DatabaseConn, err = GetDBConn()
	if err != nil {
		panic(err)
	}

	logging.L = log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Level:           log.DebugLevel,
	})

	e := echo.New()
	e.Logger = slog.New(logging.L)

	e.Renderer, err = newTemplateRenderer()
	if err != nil {
		panic(err)
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				e.Logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				e.Logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	e.Static("/static", "static")

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "html/home.html", nil)
	})

	e.GET("/login", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "html/login.html", nil)
	})

	e.GET("/download", func(c *echo.Context) error {
		clients, err := GetClients()
		if err != nil {
			c.Logger().Error("failed to query for clients", "err", err)
			return c.String(http.StatusInternalServerError, "Unknown server-side error has occured!")
		}
		return c.Render(http.StatusOK, "html/downloads.html", map[string]any{"Clients": clients})
	})

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("shutting down the server", "error", err)
	}
}
