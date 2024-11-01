package errs

import "fmt"

type WrongPasswordType struct {
	PasswordType int
}

func (e WrongPasswordType) Error() string {
	return fmt.Sprintf("Wrong password type: %d", e.PasswordType)
}
