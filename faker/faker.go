package faker

import (
	"github.com/brianvoe/gofakeit/v7"
)

func Email() string {
	return gofakeit.Email()
}

func Password() string {
	return gofakeit.Password(true, true, true, true, false, 12)
}

func Phone() string {
	return gofakeit.Phone()
}
