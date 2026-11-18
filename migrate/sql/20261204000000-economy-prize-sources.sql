-- +migrate Up
-- +migrate StatementBegin
DO $$
DECLARE
    def TEXT;
    vals TEXT[];
BEGIN
    SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
        WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
    SELECT array_agg(DISTINCT v ORDER BY v) INTO vals FROM (
        SELECT m[1] AS v FROM regexp_matches(coalesce(def, ''), '''([a-z_]+)''', 'g') AS m
        UNION SELECT unnest(string_to_array(m[1], ',')) FROM regexp_matches(coalesce(def, ''), '''\{([a-z_,]*)\}''', 'g') AS m
        UNION SELECT unnest(ARRAY['payment','gift','admin','event','level','streak','referral','task','achievement','chance','game','trophy','stage'])
    ) x WHERE v <> '';
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT IF EXISTS economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check CHECK (source_type IN (' || (SELECT string_agg(quote_literal(v), ',') FROM unnest(vals) v) || '))';
END $$;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'economy prize sources migration cannot be rolled back'; END $$;
-- +migrate StatementEnd
