package validation

import "regexp"

var phoneNumberRegex = regexp.MustCompile(`^09\d{9}$`)

func IsValidPhoneNumber(input string) bool {
	return phoneNumberRegex.MatchString(input)
}
