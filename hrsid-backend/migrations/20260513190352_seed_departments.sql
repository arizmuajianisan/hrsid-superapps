-- +goose Up
-- +goose StatementBegin
INSERT INTO departments (name) VALUES
('Engineering'),
('Production'),
('Quality Assurance'),
('Information Technology'),
('Human Resources')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM departments WHERE name IN ('Engineering', 'Production', 'Quality Assurance', 'Information Technology', 'Human Resources');
-- +goose StatementEnd
