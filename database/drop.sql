-- =========================================
-- DROP VIEWS
-- =========================================
DROP VIEW IF EXISTS vw_priority_reservations CASCADE;
DROP VIEW IF EXISTS vw_user_violations CASCADE;
DROP VIEW IF EXISTS vw_activity_usage CASCADE;
DROP VIEW IF EXISTS vw_peak_hours CASCADE;
DROP VIEW IF EXISTS vw_resource_usage CASCADE;

-- =========================================
-- DROP TABLES
-- =========================================
DROP TABLE IF EXISTS logs CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS availability_blocks CASCADE;
DROP TABLE IF EXISTS priority_reservations CASCADE;
DROP TABLE IF EXISTS violations CASCADE;
DROP TABLE IF EXISTS participants CASCADE;
DROP TABLE IF EXISTS reservations CASCADE;
DROP TABLE IF EXISTS activities CASCADE;
DROP TABLE IF EXISTS resources CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- =========================================
-- DROP FUNCTION
-- =========================================
DROP FUNCTION IF EXISTS update_timestamp() CASCADE;

-- =========================================
-- DROP EXTENSION (OPTIONAL)
-- =========================================
DROP EXTENSION IF EXISTS btree_gist CASCADE;