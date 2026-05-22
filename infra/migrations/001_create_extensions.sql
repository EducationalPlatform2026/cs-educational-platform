-- Migration 001: Enable required PostgreSQL extensions
-- UP
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
