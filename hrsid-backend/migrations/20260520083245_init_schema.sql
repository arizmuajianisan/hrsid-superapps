-- +goose Up
-- +goose StatementBegin
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    nik VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    department_id INTEGER REFERENCES departments(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash BYTEA NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_blocked BOOLEAN DEFAULT FALSE,
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

INSERT INTO departments (name) VALUES
('Engineering'), ('Production'), ('Quality Assurance'), ('Information Technology'), ('Human Resources');

INSERT INTO applications (name, slug, base_url, description) VALUES
('Hi-DSign', 'hi-dsign', 'https://hi-dsign.hrs-id.com', 'Electronic Drawing Approval System'),
('Hirose Gatepass', 'gatepass', 'https://gatepass.hrs-id.com', 'Visitor Management System');

INSERT INTO department_applications (department_id, application_id)
SELECT d.id, a.id
FROM departments d, applications a
WHERE d.name = 'Engineering' AND a.slug = 'hi-dsign';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS department_applications;
DELETE FROM departments WHERE name IN ('Engineering', 'Production', 'Quality Assurance', 'Information Technology', 'Human Resources');
-- +goose StatementEnd
