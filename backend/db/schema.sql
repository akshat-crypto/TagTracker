-- Table for storing EC2 instances
CREATE TABLE ec2_instances (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL UNIQUE,
    instance_type VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    tags JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for storing RDS instances
CREATE TABLE rds_instances (
    id SERIAL PRIMARY KEY,
    db_instance_identifier VARCHAR(255) NOT NULL UNIQUE,
    db_instance_class VARCHAR(255) NOT NULL,
    engine VARCHAR(255) NOT NULL,
    tags JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
