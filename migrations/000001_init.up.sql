CREATE TABLE users
(
    id            CHAR(36) PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255)       NOT NULL,
    role          VARCHAR(20)        NOT NULL,
    created_at    TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (role IN ('requester', 'manager'))
) ENGINE=InnoDB;

CREATE TABLE travel_requests
(
    id             CHAR(36) PRIMARY KEY,
    requester_name VARCHAR(100) NOT NULL,
    destination    VARCHAR(255) NOT NULL,
    departure_date TIMESTAMP    NOT NULL,
    return_date    TIMESTAMP    NOT NULL,
    status         VARCHAR(20)  NOT NULL,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id        CHAR(36)     NOT NULL,
    CHECK (status IN ('requested', 'approved', 'canceled')),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB;
