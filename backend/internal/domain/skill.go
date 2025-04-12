package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Skill categories
const (
	SkillCategoryLanguage  = "language"
	SkillCategoryFramework = "framework"
	SkillCategoryTool      = "tool"
	SkillCategoryDatabase  = "database"
	SkillCategoryOther     = "other"
)

// ValidSkillCategories defines valid skill categories
var ValidSkillCategories = map[string]bool{
	SkillCategoryLanguage:  true,
	SkillCategoryFramework: true,
	SkillCategoryTool:      true,
	SkillCategoryDatabase:  true,
	SkillCategoryOther:     true,
}

// Skill represents a skill entry in a resume
type Skill struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency int    `json:"proficiency,omitempty"` // 1-5 scale
}

// Validate validates the skill entry
func (s *Skill) Validate() error {
	// Validate required fields
	if strings.TrimSpace(s.Name) == "" {
		return NewValidationError("name", "Skill name is required", ErrInvalidField)
	}

	// Validate category if provided
	if s.Category != "" {
		if !ValidSkillCategories[s.Category] {
			return NewValidationError("category", fmt.Sprintf("Invalid category, must be one of: %s", strings.Join(getSkillCategoryKeys(), ", ")), ErrInvalidField)
		}
	}

	// Validate proficiency if provided
	if s.Proficiency != 0 {
		if s.Proficiency < 1 || s.Proficiency > 5 {
			return NewValidationError("proficiency", "Proficiency must be between 1 and 5", ErrInvalidField)
		}
	}

	return nil
}

// BeforeSave sanitizes the data before saving
func (s *Skill) BeforeSave() {
	s.Name = strings.TrimSpace(s.Name)
	s.Category = strings.TrimSpace(s.Category)

	// Set default category if not provided
	if s.Category == "" {
		s.Category = SkillCategoryOther
	}
}

// ToJSON converts the skill entry to JSON
func (s *Skill) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON parses the skill entry from JSON
func (s *Skill) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// Helper function to get skill category keys
func getSkillCategoryKeys() []string {
	keys := make([]string, 0, len(ValidSkillCategories))
	for k := range ValidSkillCategories {
		keys = append(keys, k)
	}
	return keys
}
