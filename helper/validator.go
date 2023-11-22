package helper

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
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
			Tag:     "bad_name",
		}
	}

	return nil
}

func ValidateEmailAddress(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return &InputValidationError{
			Message: "email address appeared to be invalid or can't be used",
			Field:   "email",
			Tag:     "bad_email",
		}
	}

	return nil
}

func ValidateShopUserName(shopUsername string) error {
	// Trim leading and trailing spaces from the shop name
	shopUsername = strings.TrimSpace(shopUsername)

	// Check if the shop name is empty
	if len(shopUsername) == 0 {
		return &InputValidationError{
			Message: "Shop username is required",
			Field:   "username",
			Tag:     "required",
		}
	}

	// Validate the shop name using regular expressions
	valid := regexp.MustCompile("^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$")
	if !valid.MatchString(shopUsername) {
		return &InputValidationError{
			Message: "Invalid shop username",
			Field:   "username",
			Tag:     "invalid",
		}
	}

	return nil
}

func ValidateShopName(shopName string) error {
	// Trim leading and trailing spaces from the shop name
	shopName = strings.TrimSpace(shopName)

	// Check if the shop name is empty
	if len(shopName) == 0 {
		return &InputValidationError{
			Message: "Shop name is required",
			Field:   "name",
			Tag:     "required",
		}
	}

	return nil
}

func ValidateShopDescription(description string) error {
	// Trim leading and trailing spaces from the description
	description = strings.TrimSpace(description)

	// Check if the description is empty
	if len(description) == 0 {
		return &InputValidationError{
			Message: "Description is required",
			Field:   "description",
			Tag:     "required",
		}
	}

	return nil
}
