-- +migrate Up notransaction
ALTER TABLE economy_category DROP CONSTRAINT IF EXISTS economy_category_kind_check;
ALTER TABLE economy_category ADD CONSTRAINT economy_category_kind_check CHECK (kind IN ('product','prize','gift','currency','faq','news'));

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Publication migration % may leave faq/news categories in use and cannot be reverted automatically; restore from backup', '20261031000000-publication-category-kinds'; END $$;
-- +migrate StatementEnd
