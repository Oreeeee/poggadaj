// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"io"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	set        *jet.Set
	translator *Translator
}

func NewTemplateRenderer(baseDir string, devMode bool, translator *Translator) (*TemplateRenderer, error) {
	renderer := &TemplateRenderer{}
	renderer.set = jet.NewSet(
		jet.NewOSFileSystemLoader(baseDir),
		jet.DevelopmentMode(devMode),
	)

	renderer.translator = translator

	renderer.set.AddGlobalFunc("t", renderer.translator.JetTranslate)
	return renderer, nil
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, passedData any) error {
	view, err := t.set.GetTemplate(name)
	if err != nil {
		return err
	}

	// Retrieve the language for i18n
	lang := "en"
	if langRaw := c.Get("poggadaj-lang"); langRaw != nil {
		lang = langRaw.(string)
	}

	return view.Execute(w, nil, TemplateArgs{
		Language: lang,
		Data:     passedData,
	})
}
