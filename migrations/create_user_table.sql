CREATE TABLE IF NOT EXISTS app.users (
    id INT NOT NULL AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    request_uuid CHAR(36) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uuid (uuid),
    UNIQUE KEY requestUUID (request_uuid),
    UNIQUE KEY name (name)
);
