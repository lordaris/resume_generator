package domain

// Resume represents a user's resume
type Resume struct {
	ID          int              `json:"id" db:"id"`
	UserID      int              `json:"user_id" db:"user_id"`
	Title       string           `json:"title" db:"title"`
	Description string           `json:"description" db:"description"`
	Skills      []string         `json:"skills"`
	Education   []Education      `json:"education"`
	WorkHistory []WorkExperience `json:"work_history"`
	CreatedAt   string           `json:"created_at" db:"created_at"`
	UpdatedAt   string           `json:"updated_at" db:"updated_at"`
}

// Education represents educational background
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

// WorkExperience represents work experience
type WorkExperience struct {
	Company     string `json:"company"`
	Position    string `json:"position"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}
