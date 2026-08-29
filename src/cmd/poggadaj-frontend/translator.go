// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/CloudyKit/jet/v6"
)

// Translator provides a mechanism to translate strings inside of templates
type Translator struct {
	i18n map[string]map[string]string
}

// JetTranslate is a function designed to be exclusively called inside of Jet.
// It translates the key provided in the second arg from the language provided in the first arg.
// If the language is invalid or the key is not provided, the key is returned.
func (t *Translator) JetTranslate(args jet.Arguments) reflect.Value {
	value := ""

	args.RequireNumOfArguments("t", 2, 2)
	langName := args.Get(0).String()
	keyName := args.Get(1).String()

	value = keyName

	lang, langOk := t.i18n[langName]
	if !langOk {
		return reflect.ValueOf(keyName)
	}

	translated, keyOk := lang[keyName]
	if !keyOk {
		return reflect.ValueOf(value)
	}

	return reflect.ValueOf(translated)
}

func (t *Translator) loadLanguages(languagesDir string) error {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.json", languagesDir))
	if err != nil {
		return err
	}

	for _, v := range files {
		file, err := os.Open(v)
		if err != nil {
			return err
		}

		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		key := ""
		/*
			_, err = fmt.Sscanf(v, languagesDir+"/%s.json", &key)
			if err != nil {
				return err
			}
		*/

		// TODO: Parse filenames here instead of hardcoding these
		switch v {
		case "i18n/en.json":
			key = "en"
		case "i18n/pl.json":
			key = "pl"
		}

		lang := map[string]string{}

		err = json.Unmarshal(data, &lang)
		if err != nil {
			return err
		}

		t.i18n[key] = lang
	}

	return nil
}

func NewTranslator(languagesDir string) (*Translator, error) {
	translator := &Translator{
		i18n: make(map[string]map[string]string),
	}

	err := translator.loadLanguages(languagesDir)
	if err != nil {
		return nil, err
	}

	return translator, nil
}
