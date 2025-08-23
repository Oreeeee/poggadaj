// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package errs

import "fmt"

type WrongPasswordType struct {
	PasswordType int
}

func (e WrongPasswordType) Error() string {
	return fmt.Sprintf("Wrong password type: %d", e.PasswordType)
}
