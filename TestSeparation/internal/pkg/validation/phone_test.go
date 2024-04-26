package validation

import "testing"

func TestIsValidPhoneNumber(t *testing.T) {
	IsValidPhoneNumber("09123456789")
	IsValidPhoneNumber("0912345678")
}
