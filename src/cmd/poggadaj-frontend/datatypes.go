// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

// TemplateArgs represents the data structure that the templates expect
type TemplateArgs struct {
	Language string // The name of the language to use for i18n
	HasAuth  bool
	Data     any // Any other arbitrary data
}
