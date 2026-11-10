CREATE SCHEMA IF NOT EXISTS zz_setup_backup;
DROP TABLE IF EXISTS zz_setup_backup.pm, zz_setup_backup.pv, zz_setup_backup.lang, zz_setup_backup.cur;
CREATE TABLE zz_setup_backup.pm AS SELECT * FROM economy_payment_method;
CREATE TABLE zz_setup_backup.pv AS SELECT * FROM publication_provider;
CREATE TABLE zz_setup_backup.lang AS SELECT * FROM base_language;
CREATE TABLE zz_setup_backup.cur AS SELECT * FROM economy_currency;
SELECT 'backed up', (SELECT count(*) FROM zz_setup_backup.pm), (SELECT count(*) FROM zz_setup_backup.pv), (SELECT count(*) FROM zz_setup_backup.lang), (SELECT count(*) FROM zz_setup_backup.cur);
