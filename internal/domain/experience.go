package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// Experience represents a work experience entry in a resume
type Experience struct {
	Employer     string   `json:"employer"`
	JobTitle     string   `json:"title"`
	Location     string   `json:"location"`
	StartDate    string   `json:"start_date"` // Format: YYYY-MM-DD
	EndDate      string   `json:"end_date"`   // Format: YYYY-MM-DD or "Present"
	Description  string   `json:"description"`
	Achievements []string `json:"achievements,omitempty"`
}

// Validate validates the work experience entry
func (e *Experience) Validate() error {
	// Validate required fields
	if strings.TrimSpace(e.Employer) == "" {
		return NewValidationError("employer", "Employer is required", ErrInvalidField)
	}
	if strings.TrimSpace(e.JobTitle) == "" {
		return NewValidationError("title", "Job title is required", ErrInvalidField)
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
func (e *Experience) BeforeSave() {
	e.Employer = strings.TrimSpace(e.Employer)
	e.JobTitle = strings.TrimSpace(e.JobTitle)
	e.Location = strings.TrimSpace(e.Location)
	e.StartDate = strings.TrimSpace(e.StartDate)
	e.EndDate = strings.TrimSpace(e.EndDate)
	e.Description = strings.TrimSpace(e.Description)

	// Trim achievements
	for i, achievement := range e.Achievements {
		e.Achievements[i] = strings.TrimSpace(achievement)
	}

	// Remove empty achievements
	filteredAchievements := make([]string, 0, len(e.Achievements))
	for _, achievement := range e.Achievements {
		if achievement != "" {
			filteredAchievements = append(filteredAchievements, achievement)
		}
	}
	e.Achievements = filteredAchievements
}

// ToJSON converts the work experience entry to JSON
func (e *Experience) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON parses the work experience entry from JSON
func (e *Experience) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}
