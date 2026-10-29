-- +migrate Up notransaction
-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'economy_gift' AND column_name = 'name' AND data_type = 'text') THEN
        SELECT code INTO def FROM base_language WHERE is_default LIMIT 1;
        IF def IS NULL THEN
            def := 'fa';
        END IF;
        EXECUTE format('ALTER TABLE economy_gift ALTER COLUMN name TYPE JSONB USING CASE WHEN btrim(name) = %L THEN %L::jsonb ELSE jsonb_build_object(%L, btrim(name)) END', '', '{}', def);
    END IF;
END $$;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Economy migration % turned gift names into translated maps and cannot be reverted automatically; restore from backup', '20261101000002-economy-gift-name-localized'; END $$;
-- +migrate StatementEnd
