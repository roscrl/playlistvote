-- Persistent PRAGMAS
PRAGMA journal_mode = WAL;    

-- FYI: Non persistent pragmas, need to be set on DB connection level on every open
PRAGMA synchronous  = OFF;   -- Do not wait for OS response when writing to disk
PRAGMA cache_size   = 50000; 
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA temp_store   = MEMORY;
PRAGMA mmap_size = 300000000;