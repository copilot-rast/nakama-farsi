-- +migrate Up
CREATE TABLE IF NOT EXISTS publication_phone_backfill_issue (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    phone TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL CHECK (reason IN ('invalid', 'phone_taken', 'user_has_other_phone')),
    owner_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate StatementBegin
DO $$
DECLARE
    found_n INT;
    moved_n INT;
    same_n INT;
    invalid_n INT;
    taken_n INT;
    other_n INT;
    stripped_n INT;
BEGIN
    CREATE TEMP TABLE phone_backfill_src ON COMMIT DROP AS
    SELECT x.id, x.create_time, x.raw,
        CASE
            WHEN x.d ~ '^\+[1-9][0-9]{7,14}$' THEN x.d
            WHEN x.d ~ '^00[1-9][0-9]{7,14}$' THEN '+' || substr(x.d, 3)
            WHEN x.d ~ '^09[0-9]{9}$' THEN '+98' || substr(x.d, 2)
            WHEN x.d ~ '^9[0-9]{9}$' THEN '+98' || x.d
            WHEN x.d ~ '^989[0-9]{9}$' THEN '+' || x.d
        END AS phone
    FROM (
        SELECT u.id, u.create_time, coalesce(u.metadata->>'phone', '') AS raw,
            translate(regexp_replace(coalesce(u.metadata->>'phone', ''), '[[:space:]()./-]', '', 'g'), '۰۱۲۳۴۵۶۷۸۹٠١٢٣٤٥٦٧٨٩', '01234567890123456789') AS d
        FROM users u
        WHERE u.metadata ? 'phone' AND u.id <> '00000000-0000-0000-0000-000000000000'
    ) x;
    SELECT count(*) INTO found_n FROM phone_backfill_src;

    SELECT count(*) INTO same_n FROM phone_backfill_src s JOIN publication_phone_account p ON p.user_id = s.id AND p.phone = s.phone;

    INSERT INTO publication_phone_backfill_issue (user_id, phone, reason)
    SELECT s.id, left(s.raw, 64), 'invalid' FROM phone_backfill_src s
    WHERE s.phone IS NULL AND NOT EXISTS (SELECT 1 FROM publication_phone_account p WHERE p.user_id = s.id)
    ON CONFLICT (user_id) DO NOTHING;
    GET DIAGNOSTICS invalid_n = ROW_COUNT;

    INSERT INTO publication_phone_backfill_issue (user_id, phone, reason, owner_id)
    SELECT s.id, s.phone, 'user_has_other_phone', s.id FROM phone_backfill_src s
    JOIN publication_phone_account p ON p.user_id = s.id
    WHERE s.phone IS NOT NULL AND p.phone <> s.phone
    ON CONFLICT (user_id) DO NOTHING;
    GET DIAGNOSTICS other_n = ROW_COUNT;

    CREATE TEMP TABLE phone_backfill_new ON COMMIT DROP AS
    SELECT s.id, s.phone, row_number() OVER (PARTITION BY s.phone ORDER BY s.create_time, s.id) AS rn
    FROM phone_backfill_src s
    WHERE s.phone IS NOT NULL
        AND NOT EXISTS (SELECT 1 FROM publication_phone_account p WHERE p.user_id = s.id)
        AND NOT EXISTS (SELECT 1 FROM publication_phone_account p WHERE p.phone = s.phone);

    INSERT INTO publication_phone_backfill_issue (user_id, phone, reason, owner_id)
    SELECT s.id, s.phone, 'phone_taken', p.user_id FROM phone_backfill_src s
    JOIN publication_phone_account p ON p.phone = s.phone AND p.user_id <> s.id
    WHERE NOT EXISTS (SELECT 1 FROM publication_phone_account q WHERE q.user_id = s.id)
    ON CONFLICT (user_id) DO NOTHING;
    GET DIAGNOSTICS taken_n = ROW_COUNT;

    INSERT INTO publication_phone_account (phone, user_id, login_key, created_at, verified_at)
    SELECT n.phone, n.id, '', now(), now() FROM phone_backfill_new n WHERE n.rn = 1
    ON CONFLICT DO NOTHING;
    GET DIAGNOSTICS moved_n = ROW_COUNT;

    INSERT INTO publication_phone_backfill_issue (user_id, phone, reason, owner_id)
    SELECT n.id, n.phone, 'phone_taken', p.user_id FROM phone_backfill_new n
    JOIN publication_phone_account p ON p.phone = n.phone
    WHERE p.user_id <> n.id
    ON CONFLICT (user_id) DO NOTHING;
    GET DIAGNOSTICS stripped_n = ROW_COUNT;
    taken_n := taken_n + stripped_n;

    UPDATE users SET metadata = metadata - 'phone' WHERE metadata ? 'phone';
    GET DIAGNOSTICS stripped_n = ROW_COUNT;

    RAISE NOTICE 'publication phone private: % users had metadata phone, % moved, % already there, % invalid, % taken by another player, % had another number, % stripped',
        found_n, moved_n, same_n, invalid_n, taken_n, other_n, stripped_n;
END $$;
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % moved phone numbers out of account metadata and cannot be reverted automatically; restore from backup', '20270201000000-publication-phone-private'; END $$;
-- +migrate StatementEnd
