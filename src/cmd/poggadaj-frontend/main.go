// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"codeberg.org/or3e/poggadaj/internal/logging"
)

func main() {
	logger := logging.NewLogger()

	translator, err := NewTranslator("i18n")
	if err != nil {
		panic(err)
	}

	renderer, err := NewTemplateRenderer("./views", true, translator)
	if err != nil {
		panic(err)
	}

	server, err := NewServer(":3000", logger, renderer)
	if err != nil {
		panic(err)
	}

	if err := server.Run(); err != nil {
		logger.Error("shutting down the server", "error", err)
	}
}
