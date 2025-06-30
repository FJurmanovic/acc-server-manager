-- Migration 002: Migrate servers and related tables from integer IDs to UUIDs
-- This migration handles: servers, configs, state_histories, steam_credentials, system_configs

PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;

-- Step 1: Create new servers table with UUID primary key
CREATE TABLE servers_new (
    id TEXT PRIMARY KEY, -- UUID stored as TEXT in SQLite
    name TEXT NOT NULL,
    ip TEXT NOT NULL,
    port INTEGER NOT NULL,
    path TEXT NOT NULL, -- Updated from config_path to path to match Go model
    service_name TEXT NOT NULL,
    date_created DATETIME,
    from_steam_cmd BOOLEAN NOT NULL DEFAULT 1 -- Added to match Go model
);

-- Step 2: Generate UUIDs for existing servers and migrate data
INSERT INTO servers_new (id, name, ip, port, path, service_name, from_steam_cmd)
SELECT
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as id,
    name,
    COALESCE(ip, '') as ip,
    COALESCE(port, 0) as port,
    COALESCE(path, '') as path,
    service_name,
    1 as from_steam_cmd
FROM servers;

-- Step 3: Create mapping table to track old ID to new UUID mapping
CREATE TEMP TABLE server_id_mapping AS
SELECT
    s_old.id as old_id,
    s_new.id as new_id
FROM servers s_old
JOIN servers_new s_new ON s_old.name = s_new.name AND s_old.service_name = s_new.service_name;

-- Step 4: Drop old servers table and rename new one
DROP TABLE servers;
ALTER TABLE servers_new RENAME TO servers;

-- Step 5: Create new configs table with UUID references
CREATE TABLE configs_new (
    id TEXT PRIMARY KEY, -- UUID for configs
    server_id TEXT NOT NULL, -- UUID reference to servers (GORM expects snake_case)
    config_file TEXT NOT NULL,
    old_config TEXT,
    new_config TEXT,
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Step 6: Migrate configs data with UUID references
INSERT INTO configs_new (id, server_id, config_file, old_config, new_config, changed_at)
SELECT
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as id,
    sim.new_id as server_id,
    c.config_file,
    c.old_config,
    c.new_config,
    c.changed_at
FROM configs c
JOIN server_id_mapping sim ON c.server_id = sim.old_id;

-- Step 7: Drop old configs table and rename new one
DROP TABLE configs;
ALTER TABLE configs_new RENAME TO configs;

-- Step 8: Create new state_histories table with UUID references
CREATE TABLE state_histories_new (
    id TEXT PRIMARY KEY, -- UUID for state_histories records
    server_id TEXT NOT NULL, -- UUID reference to servers (GORM expects snake_case)
    session TEXT,
    track TEXT,
    player_count INTEGER,
    date_created DATETIME,
    session_start DATETIME,
    session_duration_minutes INTEGER,
    session_id TEXT NOT NULL -- Changed to TEXT to store UUID
);

-- Step 9: Migrate state_histories data with UUID references
INSERT INTO state_histories_new (id, server_id, session, track, player_count, date_created, session_start, session_duration_minutes, session_id)
SELECT
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as id,
    sim.new_id as server_id,
    sh.session,
    sh.track,
    sh.player_count,
    sh.date_created,
    sh.session_start,
    sh.session_duration_minutes,
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as session_id
FROM state_histories sh
JOIN server_id_mapping sim ON sh.server_id = sim.old_id;

-- Step 10: Drop old state_histories table and rename new one
DROP TABLE state_histories;
ALTER TABLE state_histories_new RENAME TO state_histories;

-- Step 11: Create new steam_credentials table with UUID primary key
CREATE TABLE steam_credentials_new (
    id TEXT PRIMARY KEY, -- UUID for steam_credentials
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    date_created DATETIME,
    last_updated DATETIME
);

-- Step 12: Migrate steam_credentials data
INSERT INTO steam_credentials_new (id, username, password, date_created, last_updated)
SELECT
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as id,
    username,
    password,
    date_created,
    last_updated
FROM steam_credentials;

-- Step 13: Drop old steam_credentials table and rename new one
DROP TABLE steam_credentials;
ALTER TABLE steam_credentials_new RENAME TO steam_credentials;

-- Step 14: Create new system_configs table with UUID primary key
CREATE TABLE system_configs_new (
    id TEXT PRIMARY KEY, -- UUID for system_configs
    key TEXT,
    value TEXT,
    default_value TEXT,
    description TEXT,
    date_modified TEXT
);

-- Step 15: Migrate system_configs data
INSERT INTO system_configs_new (id, key, value, default_value, description, date_modified)
SELECT
    LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' ||
          SUBSTR('89AB', ABS(RANDOM()) % 4 + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6))) as id,
    key,
    value,
    default_value,
    description,
    date_modified
FROM system_configs;

-- Step 16: Drop old system_configs table and rename new one
DROP TABLE system_configs;
ALTER TABLE system_configs_new RENAME TO system_configs;

-- Step 17: Create migration record
CREATE TABLE IF NOT EXISTS migration_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration_name TEXT UNIQUE NOT NULL,
    applied_at TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    notes TEXT
);

INSERT INTO migration_records (migration_name, applied_at, success, notes)
VALUES ('002_migrate_servers_to_uuid', datetime('now'), 1, 'Migrated servers, configs, state_histories, steam_credentials, and system_configs to UUID primary keys');

COMMIT;
PRAGMA foreign_keys=ON;
