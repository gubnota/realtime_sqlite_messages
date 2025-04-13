-- schema.sql

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_seen INTEGER NOT NULL,
    score INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender TEXT NOT NULL,
    receiver TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    delivered BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    last_seen INTEGER NOT NULL,
    status CHAR(1) DEFAULT 'F',
    user_agent TEXT
);

CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender TEXT NOT NULL,
    receiver TEXT NOT NULL,
    created DATETIME NOT NULL,
    svote INTEGER DEFAULT 0,
    rvote INTEGER DEFAULT 0,
    status TEXT DEFAULT 'open' CHECK (status IN ('open', 'closed'))
);

CREATE TABLE IF NOT EXISTS results (
    user_id TEXT PRIMARY KEY,
    score INTEGER NOT NULL DEFAULT 0,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
);