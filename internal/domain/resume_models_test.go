package domain

import (
	"testing"
)

// BenchmarkPersonalInfoValidation benchmarks the validation of personal information
func BenchmarkPersonalInfoValidation(b *testing.B) {
	personalInfo := &PersonalInfo{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
		Address: struct {
			Street  string `json:"street"`
			City    string `json:"city"`
			Country string `json:"country"`
		}{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
		JobTitle: "Software Engineer",
	}

	for i := 0; i < b.N; i++ {
		personalInfo.Validate()
	}
}

// BenchmarkEducationValidation benchmarks the validation of education
func BenchmarkEducationValidation(b *testing.B) {
	education := &Education{
		Institution: "Stanford University",
		Location:    "Stanford, CA",
		Degree:      "Bachelor of Science",
		Field:       "Computer Science",
		StartDate:   "2015-09-01",
		EndDate:     "2019-06-30",
		Description: "Graduated with honors",
	}

	for i := 0; i < b.N; i++ {
		education.Validate()
	}
}

// BenchmarkExperienceValidation benchmarks the validation of experience
func BenchmarkExperienceValidation(b *testing.B) {
	experience := &Experience{
		Employer:    "Google",
		JobTitle:    "Software Engineer",
		Location:    "Mountain View, CA",
		StartDate:   "2019-07-15",
		EndDate:     "Present",
		Description: "Working on Google Cloud Platform",
		Achievements: []string{
			"Improved service reliability by 99.9%",
			"Led a team of 5 engineers",
		},
	}

	for i := 0; i < b.N; i++ {
		experience.Validate()
	}
}

// BenchmarkSkillValidation benchmarks the validation of skill
func BenchmarkSkillValidation(b *testing.B) {
	skill := &Skill{
		Name:        "Go",
		Category:    "language",
		Proficiency: 5,
	}

	for i := 0; i < b.N; i++ {
		skill.Validate()
	}
}

// BenchmarkProjectValidation benchmarks the validation of project
func BenchmarkProjectValidation(b *testing.B) {
	project := &Project{
		Name:        "Personal Website",
		Description: "A personal portfolio website built with React",
		Technologies: []string{
			"React",
			"TypeScript",
			"Tailwind CSS",
		},
		RepoURL:   "https://github.com/johndoe/personal-website",
		DemoURL:   "https://johndoe.com",
		StartDate: "2020-01-01",
		EndDate:   "2020-02-15",
	}

	for i := 0; i < b.N; i++ {
		project.Validate()
	}
}

// BenchmarkCertificationValidation benchmarks the validation of certification
func BenchmarkCertificationValidation(b *testing.B) {
	certification := &Certification{
		Name:         "AWS Certified Solutions Architect",
		Issuer:       "Amazon Web Services",
		IssueDate:    "2020-03-01",
		ExpiryDate:   "2023-03-01",
		CredentialID: "AWS-12345",
		URL:          "https://aws.amazon.com/certification/certified-solutions-architect-associate/",
	}

	for i := 0; i < b.N; i++ {
		certification.Validate()
	}
}
