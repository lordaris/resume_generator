package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/rs/zerolog/log"
)

// PostgresResumeRepository implements the ResumeRepository interface
type PostgresResumeRepository struct {
	db *sqlx.DB
}

// NewPostgresResumeRepository creates a new PostgreSQL resume repository
func NewPostgresResumeRepository(db *sqlx.DB) *PostgresResumeRepository {
	return &PostgresResumeRepository{
		db: db,
	}
}

// CreateResume creates a new resume
func (r *PostgresResumeRepository) CreateResume(userID uuid.UUID) (*domain.Resume, error) {
	query := `
		INSERT INTO resumes (id, user_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	resumeID := uuid.New()
	now := time.Now()

	var id uuid.UUID
	err := r.db.QueryRow(
		query,
		resumeID,
		userID,
		now,
	).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create resume")
		return nil, err
	}

	resume := &domain.Resume{
		ID:        resumeID,
		UserID:    userID,
		CreatedAt: now,
	}

	return resume, nil
}

// GetResumeByID retrieves a resume by ID
func (r *PostgresResumeRepository) GetResumeByID(id uuid.UUID) (*domain.Resume, error) {
	query := `
		SELECT id, user_id, created_at
		FROM resumes
		WHERE id = $1
	`

	var resume domain.Resume
	err := r.db.Get(&resume, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("resume_id", id.String()).Msg("Failed to get resume by ID")
		return nil, err
	}

	return &resume, nil
}

// GetResumesByUserID retrieves all resumes for a user
func (r *PostgresResumeRepository) GetResumesByUserID(userID uuid.UUID) ([]*domain.Resume, error) {
	query := `
		SELECT id, user_id, created_at
		FROM resumes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var resumes []*domain.Resume
	err := r.db.Select(&resumes, query, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get resumes by user ID")
		return nil, err
	}

	return resumes, nil
}

// DeleteResume deletes a resume
func (r *PostgresResumeRepository) DeleteResume(id uuid.UUID) error {
	query := `
		DELETE FROM resumes
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Str("resume_id", id.String()).Msg("Failed to delete resume")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// SavePersonalInfo saves personal info for a resume
func (r *PostgresResumeRepository) SavePersonalInfo(resumeID uuid.UUID, info *domain.PersonalInfo) error {
	query := `
		INSERT INTO personal_info (
			id, resume_id, first_name, last_name, email, phone, 
			street, city, country, job_title, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (resume_id) DO UPDATE SET
			first_name = $3,
			last_name = $4,
			email = $5,
			phone = $6,
			street = $7,
			city = $8,
			country = $9,
			job_title = $10,
			updated_at = $12
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	info.BeforeSave()

	id := uuid.New()
	now := time.Now()

	var returnedID uuid.UUID
	err := r.db.QueryRow(
		query,
		id,
		resumeID,
		info.FirstName,
		info.LastName,
		info.Email,
		info.Phone,
		info.Address.Street,
		info.Address.City,
		info.Address.Country,
		info.JobTitle,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save personal info")
		return err
	}

	return nil
}

// GetPersonalInfo retrieves personal info for a resume
func (r *PostgresResumeRepository) GetPersonalInfo(resumeID uuid.UUID) (*domain.PersonalInfo, error) {
	query := `
		SELECT first_name, last_name, email, phone, street, city, country, job_title
		FROM personal_info
		WHERE resume_id = $1
	`

	var info struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
		Phone     string `db:"phone"`
		Street    string `db:"street"`
		City      string `db:"city"`
		Country   string `db:"country"`
		JobTitle  string `db:"job_title"`
	}

	err := r.db.Get(&info, query, resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get personal info")
		return nil, err
	}

	result := &domain.PersonalInfo{
		FirstName: info.FirstName,
		LastName:  info.LastName,
		Email:     info.Email,
		Phone:     info.Phone,
		JobTitle:  info.JobTitle,
	}
	result.Address.Street = info.Street
	result.Address.City = info.City
	result.Address.Country = info.Country

	return result, nil
}

// AddEducation adds an education entry to a resume
func (r *PostgresResumeRepository) AddEducation(resumeID uuid.UUID, education *domain.Education) (uuid.UUID, error) {
	query := `
		INSERT INTO education (
			id, resume_id, institution, location, degree, field, 
			start_date, end_date, description, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	education.BeforeSave()

	// Validate the entry
	if err := education.Validate(); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if education.StartDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", education.StartDate)
		if err != nil {
			return uuid.Nil, err
		}
		startDate = &parsedStartDate
	}

	if education.EndDate != "" && education.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", education.EndDate)
		if err != nil {
			return uuid.Nil, err
		}
		endDate = &parsedEndDate
	}

	var returnedID uuid.UUID
	err := r.db.QueryRow(
		query,
		id,
		resumeID,
		education.Institution,
		education.Location,
		education.Degree,
		education.Field,
		startDate,
		endDate,
		education.Description,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add education")
		return uuid.Nil, err
	}

	return returnedID, nil
}

// UpdateEducation updates an education entry
func (r *PostgresResumeRepository) UpdateEducation(id uuid.UUID, education *domain.Education) error {
	query := `
		UPDATE education
		SET institution = $1,
			location = $2,
			degree = $3,
			field = $4,
			start_date = $5,
			end_date = $6,
			description = $7,
			updated_at = $8
		WHERE id = $9
	`

	// Apply BeforeSave to sanitize the data
	education.BeforeSave()

	// Validate the entry
	if err := education.Validate(); err != nil {
		return err
	}

	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if education.StartDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", education.StartDate)
		if err != nil {
			return err
		}
		startDate = &parsedStartDate
	}

	if education.EndDate != "" && education.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", education.EndDate)
		if err != nil {
			return err
		}
		endDate = &parsedEndDate
	}

	result, err := r.db.Exec(
		query,
		education.Institution,
		education.Location,
		education.Degree,
		education.Field,
		startDate,
		endDate,
		education.Description,
		now,
		id,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update education")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteEducation deletes an education entry
func (r *PostgresResumeRepository) DeleteEducation(id uuid.UUID) error {
	query := `
		DELETE FROM education
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete education")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetEducation retrieves an education entry by ID
func (r *PostgresResumeRepository) GetEducation(id uuid.UUID) (*domain.Education, error) {
	query := `
		SELECT institution, location, degree, field, 
		       start_date, end_date, description
		FROM education
		WHERE id = $1
	`

	var edu struct {
		Institution string     `db:"institution"`
		Location    string     `db:"location"`
		Degree      string     `db:"degree"`
		Field       string     `db:"field"`
		StartDate   time.Time  `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
		Description string     `db:"description"`
	}

	err := r.db.Get(&edu, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("education_id", id.String()).Msg("Failed to get education")
		return nil, err
	}

	// Format dates
	startDate := edu.StartDate.Format("2006-01-02")
	var endDate string
	if edu.EndDate != nil {
		endDate = edu.EndDate.Format("2006-01-02")
	} else {
		endDate = "Present"
	}

	education := &domain.Education{
		Institution: edu.Institution,
		Location:    edu.Location,
		Degree:      edu.Degree,
		Field:       edu.Field,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: edu.Description,
	}

	return education, nil
}

// GetEducationByResume retrieves all education entries for a resume
func (r *PostgresResumeRepository) GetEducationByResume(resumeID uuid.UUID) ([]*domain.Education, error) {
	query := `
		SELECT id, institution, location, degree, field, 
		       start_date, end_date, description
		FROM education
		WHERE resume_id = $1
		ORDER BY start_date DESC
	`

	type educationRow struct {
		ID          uuid.UUID  `db:"id"`
		Institution string     `db:"institution"`
		Location    string     `db:"location"`
		Degree      string     `db:"degree"`
		Field       string     `db:"field"`
		StartDate   time.Time  `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
		Description string     `db:"description"`
	}

	var rows []educationRow
	err := r.db.Select(&rows, query, resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get education by resume")
		return nil, err
	}

	education := make([]*domain.Education, len(rows))
	for i, row := range rows {
		// Format dates
		startDate := row.StartDate.Format("2006-01-02")
		var endDate string
		if row.EndDate != nil {
			endDate = row.EndDate.Format("2006-01-02")
		} else {
			endDate = "Present"
		}

		education[i] = &domain.Education{
			Institution: row.Institution,
			Location:    row.Location,
			Degree:      row.Degree,
			Field:       row.Field,
			StartDate:   startDate,
			EndDate:     endDate,
			Description: row.Description,
		}
	}

	return education, nil
}

// AddExperience adds an experience entry to a resume
func (r *PostgresResumeRepository) AddExperience(resumeID uuid.UUID, experience *domain.Experience) (uuid.UUID, error) {
	query := `
		INSERT INTO experience (
			id, resume_id, employer, job_title, location, 
			start_date, end_date, description, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	experience.BeforeSave()

	// Validate the entry
	if err := experience.Validate(); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if experience.StartDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", experience.StartDate)
		if err != nil {
			return uuid.Nil, err
		}
		startDate = &parsedStartDate
	}

	if experience.EndDate != "" && experience.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", experience.EndDate)
		if err != nil {
			return uuid.Nil, err
		}
		endDate = &parsedEndDate
	}

	var returnedID uuid.UUID
	err := r.db.QueryRow(
		query,
		id,
		resumeID,
		experience.Employer,
		experience.JobTitle,
		experience.Location,
		startDate,
		endDate,
		experience.Description,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add experience")
		return uuid.Nil, err
	}

	// Insert achievements in a transaction if there are any
	if len(experience.Achievements) > 0 {
		// TODO: Add achievements in a separate table if needed
	}

	return returnedID, nil
}

// UpdateExperience updates an experience entry
func (r *PostgresResumeRepository) UpdateExperience(id uuid.UUID, experience *domain.Experience) error {
	query := `
		UPDATE experience
		SET employer = $1,
			job_title = $2,
			location = $3,
			start_date = $4,
			end_date = $5,
			description = $6,
			updated_at = $7
		WHERE id = $8
	`

	// Apply BeforeSave to sanitize the data
	experience.BeforeSave()

	// Validate the entry
	if err := experience.Validate(); err != nil {
		return err
	}

	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if experience.StartDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", experience.StartDate)
		if err != nil {
			return err
		}
		startDate = &parsedStartDate
	}

	if experience.EndDate != "" && experience.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", experience.EndDate)
		if err != nil {
			return err
		}
		endDate = &parsedEndDate
	}

	result, err := r.db.Exec(
		query,
		experience.Employer,
		experience.JobTitle,
		experience.Location,
		startDate,
		endDate,
		experience.Description,
		now,
		id,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update experience")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Update achievements if needed
	if len(experience.Achievements) > 0 {
		// TODO: Update achievements in a separate table if needed
	}

	return nil
}

// DeleteExperience deletes an experience entry
func (r *PostgresResumeRepository) DeleteExperience(id uuid.UUID) error {
	query := `
		DELETE FROM experience
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete experience")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetExperience retrieves an experience entry by ID
func (r *PostgresResumeRepository) GetExperience(id uuid.UUID) (*domain.Experience, error) {
	query := `
		SELECT employer, job_title, location, 
		       start_date, end_date, description
		FROM experience
		WHERE id = $1
	`

	var exp struct {
		Employer    string     `db:"employer"`
		JobTitle    string     `db:"job_title"`
		Location    string     `db:"location"`
		StartDate   time.Time  `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
		Description string     `db:"description"`
	}

	err := r.db.Get(&exp, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("experience_id", id.String()).Msg("Failed to get experience")
		return nil, err
	}

	// Format dates
	startDate := exp.StartDate.Format("2006-01-02")
	var endDate string
	if exp.EndDate != nil {
		endDate = exp.EndDate.Format("2006-01-02")
	} else {
		endDate = "Present"
	}

	experience := &domain.Experience{
		Employer:    exp.Employer,
		JobTitle:    exp.JobTitle,
		Location:    exp.Location,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: exp.Description,
		// Fetch achievements if needed
		Achievements: []string{},
	}

	return experience, nil
}

// GetExperienceByResume retrieves all experience entries for a resume
func (r *PostgresResumeRepository) GetExperienceByResume(resumeID uuid.UUID) ([]*domain.Experience, error) {
	query := `
		SELECT id, employer, job_title, location, 
		       start_date, end_date, description
		FROM experience
		WHERE resume_id = $1
		ORDER BY start_date DESC
	`

	type experienceRow struct {
		ID          uuid.UUID  `db:"id"`
		Employer    string     `db:"employer"`
		JobTitle    string     `db:"job_title"`
		Location    string     `db:"location"`
		StartDate   time.Time  `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
		Description string     `db:"description"`
	}

	var rows []experienceRow
	err := r.db.Select(&rows, query, resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get experience by resume")
		return nil, err
	}

	experience := make([]*domain.Experience, len(rows))
	for i, row := range rows {
		// Format dates
		startDate := row.StartDate.Format("2006-01-02")
		var endDate string
		if row.EndDate != nil {
			endDate = row.EndDate.Format("2006-01-02")
		} else {
			endDate = "Present"
		}

		experience[i] = &domain.Experience{
			Employer:    row.Employer,
			JobTitle:    row.JobTitle,
			Location:    row.Location,
			StartDate:   startDate,
			EndDate:     endDate,
			Description: row.Description,
			// Fetch achievements if needed
			Achievements: []string{},
		}
	}

	return experience, nil
}

// AddSkill adds a skill entry to a resume
func (r *PostgresResumeRepository) AddSkill(resumeID uuid.UUID, skill *domain.Skill) (uuid.UUID, error) {
	query := `
		INSERT INTO skills (
			id, resume_id, name, category, proficiency, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	skill.BeforeSave()

	// Validate the entry
	if err := skill.Validate(); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	now := time.Now()

	// Use NULL for zero proficiency
	var proficiency any
	if skill.Proficiency != 0 {
		proficiency = skill.Proficiency
	} else {
		proficiency = nil
	}

	var returnedID uuid.UUID
	err := r.db.QueryRow(
		query,
		id,
		resumeID,
		skill.Name,
		skill.Category,
		proficiency,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add skill")
		return uuid.Nil, err
	}

	return returnedID, nil
}

// UpdateSkill updates a skill entry
func (r *PostgresResumeRepository) UpdateSkill(id uuid.UUID, skill *domain.Skill) error {
	query := `
		UPDATE skills
		SET name = $1,
			category = $2,
			proficiency = $3,
			updated_at = $4
		WHERE id = $5
	`

	// Apply BeforeSave to sanitize the data
	skill.BeforeSave()

	// Validate the entry
	if err := skill.Validate(); err != nil {
		return err
	}

	now := time.Now()

	// Use NULL for zero proficiency
	var proficiency any
	if skill.Proficiency != 0 {
		proficiency = skill.Proficiency
	} else {
		proficiency = nil
	}

	result, err := r.db.Exec(
		query,
		skill.Name,
		skill.Category,
		proficiency,
		now,
		id,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update skill")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteSkill deletes a skill entry
func (r *PostgresResumeRepository) DeleteSkill(id uuid.UUID) error {
	query := `
		DELETE FROM skills
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete skill")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetSkill retrieves a skill entry by ID
func (r *PostgresResumeRepository) GetSkill(id uuid.UUID) (*domain.Skill, error) {
	query := `
		SELECT name, category, proficiency
		FROM skills
		WHERE id = $1
	`

	var skillRow struct {
		Name        string `db:"name"`
		Category    string `db:"category"`
		Proficiency *int   `db:"proficiency"`
	}

	err := r.db.Get(&skillRow, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("skill_id", id.String()).Msg("Failed to get skill")
		return nil, err
	}

	skill := &domain.Skill{
		Name:     skillRow.Name,
		Category: skillRow.Category,
	}

	if skillRow.Proficiency != nil {
		skill.Proficiency = *skillRow.Proficiency
	}

	return skill, nil
}

// GetSkillsByResume retrieves all skill entries for a resume
func (r *PostgresResumeRepository) GetSkillsByResume(resumeID uuid.UUID) ([]*domain.Skill, error) {
	query := `
		SELECT id, name, category, proficiency
		FROM skills
		WHERE resume_id = $1
		ORDER BY category, name
	`

	type skillRow struct {
		ID          uuid.UUID `db:"id"`
		Name        string    `db:"name"`
		Category    string    `db:"category"`
		Proficiency *int      `db:"proficiency"`
	}

	var rows []skillRow
	err := r.db.Select(&rows, query, resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get skills by resume")
		return nil, err
	}

	skills := make([]*domain.Skill, len(rows))
	for i, row := range rows {
		skills[i] = &domain.Skill{
			Name:     row.Name,
			Category: row.Category,
		}

		if row.Proficiency != nil {
			skills[i].Proficiency = *row.Proficiency
		}
	}

	return skills, nil
}

// AddProject adds a project entry to a resume
func (r *PostgresResumeRepository) AddProject(resumeID uuid.UUID, project *domain.Project) (uuid.UUID, error) {
	query := `
		INSERT INTO projects (
			id, resume_id, name, description, repo_url, demo_url, 
			start_date, end_date, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	project.BeforeSave()

	// Validate the entry
	if err := project.Validate(); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if project.StartDate != "" && project.StartDate != "Present" {
		parsedStartDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			return uuid.Nil, err
		}
		startDate = &parsedStartDate
	}

	if project.EndDate != "" && project.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			return uuid.Nil, err
		}
		endDate = &parsedEndDate
	}

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return uuid.Nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var returnedID uuid.UUID
	err = tx.QueryRow(
		query,
		id,
		resumeID,
		project.Name,
		project.Description,
		project.RepoURL,
		project.DemoURL,
		startDate,
		endDate,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add project")
		return uuid.Nil, err
	}

	// Add technologies
	for _, tech := range project.Technologies {
		err = r.addProjectTechnologyTx(tx, id, tech)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add project technology")
			return uuid.Nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return uuid.Nil, err
	}

	return returnedID, nil
}

// addProjectTechnologyTx adds a technology to a project within a transaction
func (r *PostgresResumeRepository) addProjectTechnologyTx(tx *sqlx.Tx, projectID uuid.UUID, technology string) error {
	query := `
		INSERT INTO project_technologies (id, project_id, technology)
		VALUES ($1, $2, $3)
	`

	id := uuid.New()
	_, err := tx.Exec(query, id, projectID, technology)
	return err
}

// UpdateProject updates a project entry
func (r *PostgresResumeRepository) UpdateProject(id uuid.UUID, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $1,
			description = $2,
			repo_url = $3,
			demo_url = $4,
			start_date = $5,
			end_date = $6,
			updated_at = $7
		WHERE id = $8
	`

	// Apply BeforeSave to sanitize the data
	project.BeforeSave()

	// Validate the entry
	if err := project.Validate(); err != nil {
		return err
	}

	now := time.Now()

	// Parse dates
	var startDate *time.Time
	var endDate *time.Time

	if project.StartDate != "" && project.StartDate != "Present" {
		parsedStartDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			return err
		}
		startDate = &parsedStartDate
	}

	if project.EndDate != "" && project.EndDate != "Present" {
		parsedEndDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			return err
		}
		endDate = &parsedEndDate
	}

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(
		query,
		project.Name,
		project.Description,
		project.RepoURL,
		project.DemoURL,
		startDate,
		endDate,
		now,
		id,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update project")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Delete existing technologies
	_, err = tx.Exec("DELETE FROM project_technologies WHERE project_id = $1", id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete project technologies")
		return err
	}

	// Add updated technologies
	for _, tech := range project.Technologies {
		err = r.addProjectTechnologyTx(tx, id, tech)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add project technology")
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}

	return nil
}

// DeleteProject deletes a project entry
func (r *PostgresResumeRepository) DeleteProject(id uuid.UUID) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete project")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetProject retrieves a project entry by ID
func (r *PostgresResumeRepository) GetProject(id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT name, description, repo_url, demo_url, start_date, end_date
		FROM projects
		WHERE id = $1
	`

	var projectRow struct {
		Name        string     `db:"name"`
		Description string     `db:"description"`
		RepoURL     string     `db:"repo_url"`
		DemoURL     string     `db:"demo_url"`
		StartDate   *time.Time `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
	}

	err := r.db.Get(&projectRow, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("project_id", id.String()).Msg("Failed to get project")
		return nil, err
	}

	// Format dates
	var startDate, endDate string
	if projectRow.StartDate != nil {
		startDate = projectRow.StartDate.Format("2006-01-02")
	}
	if projectRow.EndDate != nil {
		endDate = projectRow.EndDate.Format("2006-01-02")
	} else {
		endDate = "Present"
	}

	// Get technologies
	technologies, err := r.GetProjectTechnologies(id)
	if err != nil {
		log.Error().Err(err).Str("project_id", id.String()).Msg("Failed to get project technologies")
		return nil, err
	}

	project := &domain.Project{
		Name:         projectRow.Name,
		Description:  projectRow.Description,
		RepoURL:      projectRow.RepoURL,
		DemoURL:      projectRow.DemoURL,
		StartDate:    startDate,
		EndDate:      endDate,
		Technologies: technologies,
	}

	return project, nil
}

// GetProjectsByResume retrieves all project entries for a resume
func (r *PostgresResumeRepository) GetProjectsByResume(resumeID uuid.UUID) ([]*domain.Project, error) {
	query := `
		SELECT id, name, description, repo_url, demo_url, start_date, end_date
		FROM projects
		WHERE resume_id = $1
		ORDER BY COALESCE(start_date, '9999-12-31') DESC
	`

	type projectRow struct {
		ID          uuid.UUID  `db:"id"`
		Name        string     `db:"name"`
		Description string     `db:"description"`
		RepoURL     string     `db:"repo_url"`
		DemoURL     string     `db:"demo_url"`
		StartDate   *time.Time `db:"start_date"`
		EndDate     *time.Time `db:"end_date"`
	}

	var rows []projectRow
	err := r.db.Select(&rows, query, resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get projects by resume")
		return nil, err
	}

	projects := make([]*domain.Project, len(rows))
	for i, row := range rows {
		// Format dates
		var startDate, endDate string
		if row.StartDate != nil {
			startDate = row.StartDate.Format("2006-01-02")
		}
		if row.EndDate != nil {
			endDate = row.EndDate.Format("2006-01-02")
		} else {
			endDate = "Present"
		}

		// Get technologies
		technologies, err := r.GetProjectTechnologies(row.ID)
		if err != nil {
			log.Error().Err(err).Str("project_id", row.ID.String()).Msg("Failed to get project technologies")
			continue
		}

		projects[i] = &domain.Project{
			Name:         row.Name,
			Description:  row.Description,
			RepoURL:      row.RepoURL,
			DemoURL:      row.DemoURL,
			StartDate:    startDate,
			EndDate:      endDate,
			Technologies: technologies,
		}
	}

	return projects, nil
}

// AddProjectTechnology adds a technology to a project
func (r *PostgresResumeRepository) AddProjectTechnology(projectID uuid.UUID, technology string) error {
	query := `
		INSERT INTO project_technologies (id, project_id, technology)
		VALUES ($1, $2, $3)
	`

	id := uuid.New()
	_, err := r.db.Exec(query, id, projectID, technology)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add project technology")
		return err
	}

	return nil
}

// DeleteProjectTechnology deletes a technology from a project
func (r *PostgresResumeRepository) DeleteProjectTechnology(projectID uuid.UUID, technology string) error {
	query := `
		DELETE FROM project_technologies
		WHERE project_id = $1 AND technology = $2
	`

	result, err := r.db.Exec(query, projectID, technology)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete project technology")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetProjectTechnologies retrieves all technologies for a project
func (r *PostgresResumeRepository) GetProjectTechnologies(projectID uuid.UUID) ([]string, error) {
	query := `
		SELECT technology
		FROM project_technologies
		WHERE project_id = $1
		ORDER BY technology
	`

	var technologies []string
	err := r.db.Select(&technologies, query, projectID)
	if err != nil {
		log.Error().Err(err).Str("project_id", projectID.String()).Msg("Failed to get project technologies")
		return nil, err
	}

	return technologies, nil
}

// AddCertification adds a certification entry to a resume
func (r *PostgresResumeRepository) AddCertification(resumeID uuid.UUID, certification *domain.Certification) (uuid.UUID, error) {
	query := `
		INSERT INTO certifications (
			id, resume_id, name, issuer, issue_date, 
			expiry_date, credential_id, url, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	// Apply BeforeSave to sanitize the data
	certification.BeforeSave()

	// Validate the entry
	if err := certification.Validate(); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	now := time.Now()

	// Parse dates
	var issueDate *time.Time
	var expiryDate *time.Time

	if certification.IssueDate != "" {
		parsedIssueDate, err := time.Parse("2006-01-02", certification.IssueDate)
		if err != nil {
			return uuid.Nil, err
		}
		issueDate = &parsedIssueDate
	}

	if certification.ExpiryDate != "" && certification.ExpiryDate != "No Expiration" {
		parsedExpiryDate, err := time.Parse("2006-01-02", certification.ExpiryDate)
		if err != nil {
			return uuid.Nil, err
		}
		expiryDate = &parsedExpiryDate
	}

	var returnedID uuid.UUID
	err := r.db.QueryRow(
		query,
		id,
		resumeID,
		certification.Name,
		certification.Issuer,
		issueDate,
		expiryDate,
		certification.CredentialID,
		certification.URL,
		now,
		now,
	).Scan(&returnedID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add certification")
		return uuid.Nil, err
	}

	return returnedID, nil
}

// UpdateCertification updates a certification entry
func (r *PostgresResumeRepository) UpdateCertification(id uuid.UUID, certification *domain.Certification) error {
	query := `
		UPDATE certifications
		SET name = $1,
			issuer = $2,
			issue_date = $3,
			expiry_date = $4,
			credential_id = $5,
			url = $6,
			updated_at = $7
		WHERE id = $8
	`

	// Apply BeforeSave to sanitize the data
	certification.BeforeSave()

	// Validate the entry
	if err := certification.Validate(); err != nil {
		return err
	}

	now := time.Now()

	// Parse dates
	var issueDate *time.Time
	var expiryDate *time.Time

	if certification.IssueDate != "" {
		parsedIssueDate, err := time.Parse("2006-01-02", certification.IssueDate)
		if err != nil {
			return err
		}
		issueDate = &parsedIssueDate
	}

	if certification.ExpiryDate != "" && certification.ExpiryDate != "No Expiration" {
		parsedExpiryDate, err := time.Parse("2006-01-02", certification.ExpiryDate)
		if err != nil {
			return err
		}
		expiryDate = &parsedExpiryDate
	}

	result, err := r.db.Exec(
		query,
		certification.Name,
		certification.Issuer,
		issueDate,
		expiryDate,
		certification.CredentialID,
		certification.URL,
		now,
		id,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update certification")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteCertification deletes a certification entry
func (r *PostgresResumeRepository) DeleteCertification(id uuid.UUID) error {
	query := `
		DELETE FROM certifications
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete certification")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetCertification retrieves a certification entry by ID
func (r *PostgresResumeRepository) GetCertification(id uuid.UUID) (*domain.Certification, error) {
	query := `
		SELECT name, issuer, issue_date, expiry_date, credential_id, url
		FROM certifications
		WHERE id = $1
	`

	var certRow struct {
		Name         string     `db:"name"`
		Issuer       string     `db:"issuer"`
		IssueDate    time.Time  `db:"issue_date"`
		ExpiryDate   *time.Time `db:"expiry_date"`
		CredentialID string     `db:"credential_id"`
		URL          string     `db:"url"`
	}

	err := r.db.Get(&certRow, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("certification_id", id.String()).Msg("Failed to get certification")
		return nil, err
	}

	// Format dates
	issueDate := certRow.IssueDate.Format("2006-01-02")
	var expiryDate string
	if certRow.ExpiryDate != nil {
		expiryDate = certRow.ExpiryDate.Format("2006-01-02")
	} else {
		expiryDate = "No Expiration"
	}

	certification := &domain.Certification{
		Name:         certRow.Name,
		Issuer:       certRow.Issuer,
		IssueDate:    issueDate,
		ExpiryDate:   expiryDate,
		CredentialID: certRow.CredentialID,
		URL:          certRow.URL,
	}

	return certification, nil
}

// GetCertificationsByResume retrieves all certification entries for a resume
func (r *PostgresResumeRepository) GetCertificationsByResume(resumeID uuid.UUID) ([]*domain.Certification, error) {
	query := `
		SELECT id, name, issuer, issue_date, expiry_date, credential_id, url
		FROM certifications
		WHERE resume_id = $1
		ORDER BY issue_date DESC
	`

	type certRow struct {
		ID           uuid.UUID  `db:"id"`
		Name         string     `db:"name"`
		Issuer       string     `db:"issuer"`
		IssueDate    time.Time  `db:"issue_date"`
		ExpiryDate   *time.Time `db:"expiry_date"`
		CredentialID string     `db:"credential_id"`
		URL          string     `db:"url"`
	}

	var rows []certRow
	err := r.db.Select(&rows, query, resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get certifications by resume")
		return nil, err
	}

	certifications := make([]*domain.Certification, len(rows))
	for i, row := range rows {
		// Format dates
		issueDate := row.IssueDate.Format("2006-01-02")
		var expiryDate string
		if row.ExpiryDate != nil {
			expiryDate = row.ExpiryDate.Format("2006-01-02")
		} else {
			expiryDate = "No Expiration"
		}

		certifications[i] = &domain.Certification{
			Name:         row.Name,
			Issuer:       row.Issuer,
			IssueDate:    issueDate,
			ExpiryDate:   expiryDate,
			CredentialID: row.CredentialID,
			URL:          row.URL,
		}
	}

	return certifications, nil
}

// GetCompleteResume retrieves a resume with all its sections
func (r *PostgresResumeRepository) GetCompleteResume(resumeID uuid.UUID) (*domain.Resume, error) {
	// Get basic resume info
	resume, err := r.GetResumeByID(resumeID)
	if err != nil {
		return nil, err
	}

	// Get personal info
	personalInfo, err := r.GetPersonalInfo(resumeID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if personalInfo != nil {
		resume.PersonalInfo = personalInfo
	}

	// Get education entries
	education, err := r.GetEducationByResume(resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get education")
	} else {
		resume.Education = education
	}

	// Get experience entries
	experience, err := r.GetExperienceByResume(resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get experience")
	} else {
		resume.Experience = experience
	}

	// Get skills
	skills, err := r.GetSkillsByResume(resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get skills")
	} else {
		resume.Skills = skills
	}

	// Get projects
	projects, err := r.GetProjectsByResume(resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get projects")
	} else {
		resume.Projects = projects
	}

	// Get certifications
	certifications, err := r.GetCertificationsByResume(resumeID)
	if err != nil {
		log.Error().Err(err).Str("resume_id", resumeID.String()).Msg("Failed to get certifications")
	} else {
		resume.Certifications = certifications
	}

	return resume, nil
}
