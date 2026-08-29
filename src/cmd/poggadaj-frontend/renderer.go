// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"io"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	set  *jet.Set
	i18n map[string]*map[string]string
}

func NewTemplateRenderer(baseDir string, devMode bool) (*TemplateRenderer, error) {
	renderer := &TemplateRenderer{}
	renderer.set = jet.NewSet(
		jet.NewOSFileSystemLoader(baseDir),
		jet.DevelopmentMode(devMode),
	)
	return renderer, nil
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, passedData any) error {
	view, err := t.set.GetTemplate(name)
	if err != nil {
		return err
	}
	return view.Execute(w, nil, passedData)
}
