-- Deliberately remove all Jazebe schema and data from a Nakama database.
--
-- Jazebe migrations refuse automatic "migrate down" because they hold user
-- data. Use this only after taking and verifying a backup:
--
--   pg_dump -Fc -d <db> -f nakama-before-jazebe-rollback.dump
--   psql -d <db> -v ON_ERROR_STOP=1 -v jazebe_confirm=DROP_JAZEBE_DATA -f scripts/jazebe-rollback.sql
--
-- Nakama core tables are not touched (none references a Jazebe table). The
-- matching migration_info rows are removed in the same transaction, so a
-- later "nakama migrate up" recreates the Jazebe schema from scratch.

\set ON_ERROR_STOP 1
\if :{?jazebe_confirm}
\else
  \echo 'Refusing: pass -v jazebe_confirm=DROP_JAZEBE_DATA after taking a verified backup.'
  \quit
\endif
SELECT :'jazebe_confirm' = 'DROP_JAZEBE_DATA' AS jazebe_confirmed \gset
\if :jazebe_confirmed
\else
  \echo 'Refusing: jazebe_confirm must be DROP_JAZEBE_DATA.'
  \quit
\endif

BEGIN;
DO $$
DECLARE t text;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'jazebe\_%' LOOP
        EXECUTE format('DROP TABLE IF EXISTS public.%I CASCADE', t);
    END LOOP;
END $$;
DELETE FROM migration_info WHERE id LIKE '%-jazebe-%';
COMMIT;
\echo 'Jazebe schema removed. Run "nakama migrate up" to recreate it.'
