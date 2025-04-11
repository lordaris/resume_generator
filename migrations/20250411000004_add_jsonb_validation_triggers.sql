-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Create validation functions for each section type

-- Personal section validation
CREATE OR REPLACE FUNCTION validate_personal_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'personal' THEN
        -- Check required fields
        IF NOT (NEW.personal_data ? 'first_name' AND 
                NEW.personal_data ? 'last_name' AND 
                NEW.personal_data ? 'email') THEN
            RAISE EXCEPTION 'Personal section must contain first_name, last_name, and email';
        END IF;
        
        -- Validate email format
        IF NEW.personal_data->>'email' !~ '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$' THEN
            RAISE EXCEPTION 'Invalid email format in personal section';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Education section validation
CREATE OR REPLACE FUNCTION validate_education_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'education' THEN
        -- Check required fields
        IF NOT (NEW.education_data ? 'institution' AND 
                NEW.education_data ? 'degree') THEN
            RAISE EXCEPTION 'Education section must contain institution and degree';
        END IF;
        
        -- Validate dates if present
        IF NEW.education_data ? 'start_date' AND NEW.education_data->>'start_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid start_date format in education section (should be YYYY-MM-DD)';
        END IF;
        
        IF NEW.education_data ? 'end_date' AND NEW.education_data->>'end_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid end_date format in education section (should be YYYY-MM-DD)';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Experience section validation
CREATE OR REPLACE FUNCTION validate_experience_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'experience' THEN
        -- Check required fields
        IF NOT (NEW.experience_data ? 'employer' AND 
                NEW.experience_data ? 'title') THEN
            RAISE EXCEPTION 'Experience section must contain employer and title';
        END IF;
        
        -- Validate dates if present
        IF NEW.experience_data ? 'start_date' AND NEW.experience_data->>'start_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid start_date format in experience section (should be YYYY-MM-DD)';
        END IF;
        
        IF NEW.experience_data ? 'end_date' AND NEW.experience_data->>'end_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid end_date format in experience section (should be YYYY-MM-DD)';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Skill section validation
CREATE OR REPLACE FUNCTION validate_skill_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'skill' THEN
        -- Check required fields
        IF NOT (NEW.skill_data ? 'name') THEN
            RAISE EXCEPTION 'Skill section must contain name';
        END IF;
        
        -- Validate category if present
        IF NEW.skill_data ? 'category' AND NOT (NEW.skill_data->>'category' IN ('language', 'framework', 'tool', 'database')) THEN
            RAISE EXCEPTION 'Skill category must be one of: language, framework, tool, database';
        END IF;
        
        -- Validate proficiency if present
        IF NEW.skill_data ? 'proficiency' THEN
            BEGIN
                IF (NEW.skill_data->>'proficiency')::INT < 1 OR (NEW.skill_data->>'proficiency')::INT > 5 THEN
                    RAISE EXCEPTION 'Skill proficiency must be between 1 and 5';
                END IF;
            EXCEPTION
                WHEN others THEN
                    RAISE EXCEPTION 'Skill proficiency must be a valid integer';
            END;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Project section validation
CREATE OR REPLACE FUNCTION validate_project_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'project' THEN
        -- Check required fields
        IF NOT (NEW.project_data ? 'name') THEN
            RAISE EXCEPTION 'Project section must contain name';
        END IF;
        
        -- Validate technologies as array if present
        IF NEW.project_data ? 'technologies' AND jsonb_typeof(NEW.project_data->'technologies') != 'array' THEN
            RAISE EXCEPTION 'Project technologies must be an array';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Certification section validation
CREATE OR REPLACE FUNCTION validate_certification_section() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type = 'certification' THEN
        -- Check required fields
        IF NOT (NEW.certification_data ? 'name' AND 
                NEW.certification_data ? 'issuer') THEN
            RAISE EXCEPTION 'Certification section must contain name and issuer';
        END IF;
        
        -- Validate dates if present
        IF NEW.certification_data ? 'issue_date' AND NEW.certification_data->>'issue_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid issue_date format in certification section (should be YYYY-MM-DD)';
        END IF;
        
        IF NEW.certification_data ? 'expiry_date' AND NEW.certification_data->>'expiry_date' !~ '^\d{4}-\d{2}-\d{2}$' THEN
            RAISE EXCEPTION 'Invalid expiry_date format in certification section (should be YYYY-MM-DD)';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for validation
CREATE TRIGGER validate_personal_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_personal_section();

CREATE TRIGGER validate_education_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_education_section();

CREATE TRIGGER validate_experience_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_experience_section();

CREATE TRIGGER validate_skill_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_skill_section();

CREATE TRIGGER validate_project_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_project_section();

CREATE TRIGGER validate_certification_section_trigger
    BEFORE INSERT OR UPDATE ON resume_sections
    FOR EACH ROW
    EXECUTE FUNCTION validate_certification_section();

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TRIGGER IF EXISTS validate_certification_section_trigger ON resume_sections;
DROP TRIGGER IF EXISTS validate_project_section_trigger ON resume_sections;
DROP TRIGGER IF EXISTS validate_skill_section_trigger ON resume_sections;
DROP TRIGGER IF EXISTS validate_experience_section_trigger ON resume_sections;
DROP TRIGGER IF EXISTS validate_education_section_trigger ON resume_sections;
DROP TRIGGER IF EXISTS validate_personal_section_trigger ON resume_sections;

DROP FUNCTION IF EXISTS validate_certification_section();
DROP FUNCTION IF EXISTS validate_project_section();
DROP FUNCTION IF EXISTS validate_skill_section();
DROP FUNCTION IF EXISTS validate_experience_section();
DROP FUNCTION IF EXISTS validate_education_section();
DROP FUNCTION IF EXISTS validate_personal_section();
