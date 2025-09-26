-- Create users table
CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  username VARCHAR(30) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  role user_role NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE UNIQUE INDEX idx_unique_admin_email ON users (email) WHERE role = 'admin';
CREATE UNIQUE INDEX idx_unique_user_email ON users (email) WHERE role = 'user';
