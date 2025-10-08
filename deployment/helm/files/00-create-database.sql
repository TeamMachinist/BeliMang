-- Ensure database exists (this runs as postgres user)
SELECT 'CREATE DATABASE belimang' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'belimang')\gexec