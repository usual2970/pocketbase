-- Initialize database for PocketBase testing
-- This script runs when the MySQL container starts for the first time

-- Create the database if it doesn't exist (already created by MYSQL_DATABASE env var)
-- CREATE DATABASE IF NOT EXISTS pocketbase;

-- Set character set and collation for better compatibility
ALTER DATABASE pocketbase CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Grant necessary privileges to the pocketbase user
GRANT ALL PRIVILEGES ON pocketbase.* TO 'pocketbase'@'%';
FLUSH PRIVILEGES;
