-- +migrate Up
ALTER TABLE publication_notification ADD COLUMN IF NOT EXISTS prize_id UUID REFERENCES economy_prize (id) ON DELETE SET NULL;
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ;
-- +migrate StatementBegin
DO $$
DECLARE def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO def FROM pg_constraint
  WHERE conname = 'economy_user_prize_source_type_check' AND conrelid = 'economy_user_prize'::regclass;
  IF def IS NOT NULL AND position('''mail''' IN def) = 0 THEN
    EXECUTE 'ALTER TABLE economy_user_prize DROP CONSTRAINT economy_user_prize_source_type_check';
    EXECUTE 'ALTER TABLE economy_user_prize ADD CONSTRAINT economy_user_prize_source_type_check ' || replace(def, 'ARRAY[', 'ARRAY[''mail''::text, ');
  END IF;
END $$;
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE publication_notification_recipient DROP COLUMN IF EXISTS claimed_at;
ALTER TABLE publication_notification DROP COLUMN IF EXISTS prize_id;
