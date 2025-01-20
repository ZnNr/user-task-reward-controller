CREATE TABLE IF NOT EXISTS users
(
    user_id SERIAL PRIMARY KEY not null,
    username VARCHAR(255) not null unique,
    email VARCHAR(255) UNIQUE,
    balance INT not null DEFAULT 0,
    password VARCHAR(255) not null,
    refer_from VARCHAR(255) DEFAULT null,
    refer_code VARCHAR(255) DEFAULT null
);

CREATE TABLE IF NOT EXISTS tasks
(
    task_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) DEFAULT null,
    price int DEFAULT 1
);

CREATE TABLE IF NOT EXISTS task_complete
(
    id SERIAL PRIMARY KEY,
    user_id int references users (user_id) on delete cascade not null,
    task_id int references tasks (task_id) on delete cascade not null
);