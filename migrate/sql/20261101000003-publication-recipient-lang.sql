-- +migrate Up notransaction
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS lang TEXT;
ALTER TABLE publication_notification_recipient ADD COLUMN IF NOT EXISTS message TEXT;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % records the delivered language per recipient and cannot be reverted automatically; restore from backup', '20261101000003-publication-recipient-lang'; END $$;
-- +migrate StatementEnd
