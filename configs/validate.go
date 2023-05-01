package configs

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

type InputValidationError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Tag     string `json:"tag"`
}

func (err *InputValidationError) Error() string {
	return err.Message
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", errors.New("unable to hash and encrypt password")
	}

	return string(bytes), nil
}

func CheckPassword(currentPassword, givenPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(givenPassword))
	if err != nil {
		return err
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &InputValidationError{
			Message: "Password should be of 8 characters long",
			Field:   "password",
			Tag:     "strong_password",
		}
	}

	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return &InputValidationError{
			Message: "Password should contain at least one lower case character",
			Field:   "password",
			Tag:     "strong_password",
		}
	}

	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return &InputValidationError{
			Message: "Password should contain at least one upper case character",
			Field:   "password",
			Tag:     "strong_password",
		}
	}

	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		return err
	}
	if !done {
		return &InputValidationError{
			Message: "Password should contain at least one digit",
			Field:   "password",
			Tag:     "strong_password",
		}
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		return err
	}
	if !done {
		return &InputValidationError{
			Message: "Password should contain at least one special character",
			Field:   "password",
			Tag:     "strong_password",
		}
	}

	return nil
}

func ValidateLoginName(name string) error {
	done, err := regexp.MatchString("^[A-Za-z][A-Za-z0-9_]{7,29}$", name)
	if err != nil {
		return err
	}

	if !done {
		return &InputValidationError{
			Message: "Login name appeared to be invalid or can't be used",
			Field:   "login_name",
			Tag:     "bad_password",
		}
	}

	return nil
}
