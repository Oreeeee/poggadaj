// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"io"
	"os"
	"reflect"
	"time"

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
	renderer.set.AddGlobalFunc("formatDate", func(args jet.Arguments) reflect.Value {
		args.RequireNumOfArguments("formatDate", 1, 1)
		dateArg := args.Get(0)

		date, ok := dateArg.Interface().(time.Time)
		if !ok {
			// The arg wasn't a time.Time
			return reflect.ValueOf(dateArg)
		}

		return reflect.ValueOf(date.Format("2006-01-02 15:04:05"))
	})
	renderer.set.AddGlobal("adminEmail", os.Getenv("GG_INSTANCE_ADMIN_EMAIL"))
	return renderer, nil
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, passedData any) error {
	view, err := t.set.GetTemplate(name)
	if err != nil {
		return err
	}

	// Retrieve the language for i18n
	lang := GetLang(c)
	hasAuth := GetHasAuth(c)

	return view.Execute(w, nil, TemplateArgs{
		Language: lang,
		HasAuth:  hasAuth,
		Data:     passedData,
	})
}
