// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package logging

import (
	"os"
	"time"

	"charm.land/log/v2"
)

func NewLoggerWithPrefix(prefix string) *log.Logger {
	logger := NewLogger()
	logger.SetPrefix(prefix)
	return logger
}

func NewLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Level:           log.DebugLevel,
	})
}
