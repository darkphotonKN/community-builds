-- Build Skills Table
-- Join Table between Skills and Builds

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS build_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS build_skills;
-- +goose StatementEnd