package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// Education represents an education entry in a resume
type Education struct {
	Institution string `json:"institution"`
	Location    string `json:"location"`
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	StartDate   string `json:"start_date"` // Format: YYYY-MM-DD
	EndDate     string `json:"end_date"`   // Format: YYYY-MM-DD or "Present"
	Description string `json:"description"`
}

// Validate validates the education entry
func (e *Education) Validate() error {
	// Validate required fields
	if strings.TrimSpace(e.Institution) == "" {
		return NewValidationError("institution", "Institution is required", ErrInvalidField)
	}
	if strings.TrimSpace(e.Degree) == "" {
		return NewValidationError("degree", "Degree is required", ErrInvalidField)
	}
	if strings.TrimSpace(e.StartDate) == "" {
		return NewValidationError("start_date", "Start date is required", ErrInvalidField)
	}

	// Validate dates
	if e.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", e.StartDate)
		if err != nil {
			return NewValidationError("start_date", "Invalid start date format (must be YYYY-MM-DD)", ErrInvalidField)
		}

		// Validate end date if provided and not "Present"
		if e.EndDate != "" && e.EndDate != "Present" {
			endDate, err := time.Parse("2006-01-02", e.EndDate)
			if err != nil {
				return NewValidationError("end_date", "Invalid end date format (must be YYYY-MM-DD or 'Present')", ErrInvalidField)
			}

			// End date must be after start date
			if endDate.Before(startDate) {
				return NewValidationError("end_date", "End date must be after start date", ErrDateRange)
			}
		}
	}

	return nil
}

// BeforeSave sanitizes the data before saving
func (e *Education) BeforeSave() {
	e.Institution = strings.TrimSpace(e.Institution)
	e.Location = strings.TrimSpace(e.Location)
	e.Degree = strings.TrimSpace(e.Degree)
	e.Field = strings.TrimSpace(e.Field)
	e.StartDate = strings.TrimSpace(e.StartDate)
	e.EndDate = strings.TrimSpace(e.EndDate)
	e.Description = strings.TrimSpace(e.Description)
}

// ToJSON converts the education entry to JSON
func (e *Education) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON parses the education entry from JSON
func (e *Education) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}
