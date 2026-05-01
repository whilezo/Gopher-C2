CREATE TABLE IF NOT EXISTS implants (
    id TEXT PRIMARY KEY,
    ip_address TEXT,
    last_seen DATETIME,
    created_at DATETIME
)