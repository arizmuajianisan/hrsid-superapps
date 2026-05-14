-- +goose Up
-- +goose StatementBegin
CREATE TABLE applications (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL, -- Contoh: 'hi-dsign', 'gatepass'
    base_url VARCHAR(255) NOT NULL,    -- URL aplikasi tujuan
    icon_url TEXT,                     -- URL icon untuk di dashboard
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabel Pivot: Menentukan aplikasi apa yang bisa diakses departemen apa
CREATE TABLE department_applications (
    department_id INTEGER REFERENCES departments(id) ON DELETE CASCADE,
    application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
    PRIMARY KEY (department_id, application_id)
);

INSERT INTO applications (name, slug, base_url, description) VALUES
('Hi-DSign', 'hi-dsign', 'https://hi-dsign.hrs-id.com', 'Electronic Drawing Approval System'),
('Hirose Gatepass', 'gatepass', 'https://gatepass.hrs-id.com', 'Visitor Management System');

-- Contoh: Berikan akses Engineering (ID 1) ke Hi-DSign (ID 1)
INSERT INTO department_applications (department_id, application_id) VALUES (1, 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS department_applications;
DROP TABLE IF EXISTS applications;
-- +goose StatementEnd
