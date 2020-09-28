package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type valFunc func(data interface{}) error

// Validate takes in some data that needs to be validated as well as a number
// of options for what validation is required
func Validate(data interface{}, vals ...valFunc) error {
	for _, val := range vals {
		// Run each validation option, returning error if one is found
		if err := val(data); err != nil {
			return err
		}
	}
	// If no errors found return nil
	return nil
}

// Checks to see if input has been entered
func isRequired(input interface{}) error {
	switch data := input.(type) {
	case string:
		input = strings.TrimSpace(data)
	}

	if input == "" {
		return errors.New("Field is required")
	}
	return nil
}

func isGreaterThan0(input interface{}) error {
	var num int
	switch x := input.(type) {
	case float64:
		num = int(x)
	case int:
		num = x
	}

	if num <= 0 {
		return errors.New("Cannot be less than or equal to zero")
	}
	return nil
}

func isGreaterThan(n int) valFunc {
	return valFunc(func(input interface{}) error {
		num := input.(int)
		if num <= n {
			return fmt.Errorf("Value not big enough, must be greater than %d", n)
		}
		return nil
	})
}

func isEmailFormat(input interface{}) error {
	email, ok := input.(string)
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
	if !emailRegex.MatchString(email) || !ok {
		return errors.New("Not a valid email format")
	}
	return nil
}

func isMinLength(minLength int) valFunc {
	return valFunc(func(input interface{}) error {
		data := input.(string)
		if data == "" {
			return nil
		}
		if len(data) < minLength {
			return fmt.Errorf("Not long enough, must be %d characters\n", minLength)
		}
		return nil
	})
}
