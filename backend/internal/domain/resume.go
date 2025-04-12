package domain

import (
	"time"

	"github.com/google/uuid"
)

// Resume represents a resume in the system
type Resume struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Optional fields not stored in the resume table
	PersonalInfo   *PersonalInfo    `json:"personal_info,omitempty" db:"-"`
	Education      []*Education     `json:"education,omitempty" db:"-"`
	Experience     []*Experience    `json:"experience,omitempty" db:"-"`
	Skills         []*Skill         `json:"skills,omitempty" db:"-"`
	Projects       []*Project       `json:"projects,omitempty" db:"-"`
	Certifications []*Certification `json:"certifications,omitempty" db:"-"`
}

// ResumeRepository defines the interface for resume data operations
type ResumeRepository interface {
	// Resume operations
	CreateResume(userID uuid.UUID) (*Resume, error)
	GetResumeByID(id uuid.UUID) (*Resume, error)
	GetResumesByUserID(userID uuid.UUID) ([]*Resume, error)
	DeleteResume(id uuid.UUID) error

	// Personal info operations
	SavePersonalInfo(resumeID uuid.UUID, info *PersonalInfo) error
	GetPersonalInfo(resumeID uuid.UUID) (*PersonalInfo, error)

	// Education operations
	AddEducation(resumeID uuid.UUID, education *Education) (uuid.UUID, error)
	UpdateEducation(id uuid.UUID, education *Education) error
	DeleteEducation(id uuid.UUID) error
	GetEducation(id uuid.UUID) (*Education, error)
	GetEducationByResume(resumeID uuid.UUID) ([]*Education, error)

	// Experience operations
	AddExperience(resumeID uuid.UUID, experience *Experience) (uuid.UUID, error)
	UpdateExperience(id uuid.UUID, experience *Experience) error
	DeleteExperience(id uuid.UUID) error
	GetExperience(id uuid.UUID) (*Experience, error)
	GetExperienceByResume(resumeID uuid.UUID) ([]*Experience, error)

	// Skill operations
	AddSkill(resumeID uuid.UUID, skill *Skill) (uuid.UUID, error)
	UpdateSkill(id uuid.UUID, skill *Skill) error
	DeleteSkill(id uuid.UUID) error
	GetSkill(id uuid.UUID) (*Skill, error)
	GetSkillsByResume(resumeID uuid.UUID) ([]*Skill, error)

	// Project operations
	AddProject(resumeID uuid.UUID, project *Project) (uuid.UUID, error)
	UpdateProject(id uuid.UUID, project *Project) error
	DeleteProject(id uuid.UUID) error
	GetProject(id uuid.UUID) (*Project, error)
	GetProjectsByResume(resumeID uuid.UUID) ([]*Project, error)

	// Project technologies operations
	AddProjectTechnology(projectID uuid.UUID, technology string) error
	DeleteProjectTechnology(projectID uuid.UUID, technology string) error
	GetProjectTechnologies(projectID uuid.UUID) ([]string, error)

	// Certification operations
	AddCertification(resumeID uuid.UUID, certification *Certification) (uuid.UUID, error)
	UpdateCertification(id uuid.UUID, certification *Certification) error
	DeleteCertification(id uuid.UUID) error
	GetCertification(id uuid.UUID) (*Certification, error)
	GetCertificationsByResume(resumeID uuid.UUID) ([]*Certification, error)

	// Complete resume operations
	GetCompleteResume(resumeID uuid.UUID) (*Resume, error)
}
