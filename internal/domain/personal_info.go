package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

// EmailRegex is the regular expression for validating email addresses (RFC 5322)
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// PhoneRegex is the regular expression for validating phone numbers (E.164 format)
var PhoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// PersonalInfo represents the personal information section of a resume
type PersonalInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"address"`
	JobTitle string `json:"job_title"`
}

// Validate validates the personal information
func (p *PersonalInfo) Validate() error {
	// Validate required fields
	if strings.TrimSpace(p.FirstName) == "" {
		return NewValidationError("first_name", "First name is required", ErrInvalidField)
	}
	if strings.TrimSpace(p.LastName) == "" {
		return NewValidationError("last_name", "Last name is required", ErrInvalidField)
	}
	if strings.TrimSpace(p.Email) == "" {
		return NewValidationError("email", "Email is required", ErrInvalidField)
	}

	// Validate email format
	if !EmailRegex.MatchString(p.Email) {
		return NewValidationError("email", "Invalid email format", ErrInvalidField)
	}

	// Validate phone format if provided
	if p.Phone != "" && !PhoneRegex.MatchString(p.Phone) {
		return NewValidationError("phone", "Invalid phone format (must be E.164 format, e.g., +1234567890)", ErrInvalidField)
	}

	// Validate address if provided
	if p.Address.Street != "" || p.Address.City != "" || p.Address.Country != "" {
		if strings.TrimSpace(p.Address.Street) == "" {
			return NewValidationError("address.street", "Street is required if address is provided", ErrInvalidField)
		}
		if strings.TrimSpace(p.Address.City) == "" {
			return NewValidationError("address.city", "City is required if address is provided", ErrInvalidField)
		}
		if strings.TrimSpace(p.Address.Country) == "" {
			return NewValidationError("address.country", "Country is required if address is provided", ErrInvalidField)
		}
	}

	return nil
}

// BeforeSave sanitizes the data before saving
func (p *PersonalInfo) BeforeSave() {
	p.FirstName = strings.TrimSpace(p.FirstName)
	p.LastName = strings.TrimSpace(p.LastName)
	p.Email = strings.TrimSpace(p.Email)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Address.Street = strings.TrimSpace(p.Address.Street)
	p.Address.City = strings.TrimSpace(p.Address.City)
	p.Address.Country = strings.TrimSpace(p.Address.Country)
	p.JobTitle = strings.TrimSpace(p.JobTitle)
}

// ToJSON converts the personal information to JSON
func (p *PersonalInfo) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON parses the personal information from JSON
func (p *PersonalInfo) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}
