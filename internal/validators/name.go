package validators

import (
	"errors"
	"regexp"
)

// ValidateNameFormat checks if the provided name follows the required naming rule.
func ValidateNameFormat(name string) error {
	validName, err := regexp.MatchString("([A-Z][a-zA-Z]*)", name)
	if err != nil {
		return err
	}
	if !validName {
		return errors.New("name should follow the naming rule")
	}
	return nil
}