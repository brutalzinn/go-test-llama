
-- create_users_table.sql

CREATE TABLE IF NOT EXISTS Users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    status ENUM('active', 'pending', 'closed') NOT NULL DEFAULT 'pending'
);

-- Insert some initial users (seed data)
INSERT INTO Users (username, password, status) VALUES
    ('john_doe', 'password123', 'active'),
    ('jane_smith', 'securepass', 'pending'),
    ('peter_jones', 'pass@word', 'closed'),
    ('alice_wonderland', 'secret!', 'active');