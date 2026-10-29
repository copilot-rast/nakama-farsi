-- +migrate Up notransaction
-- +migrate StatementBegin
DO $$
BEGIN
    IF to_regclass('economy_category') IS NOT NULL THEN
        INSERT INTO base_category (id, key_type, key, name, sort, created_at, updated_at)
        SELECT id, kind, key, name, sort, created_at, updated_at FROM economy_category
        ON CONFLICT DO NOTHING;
    END IF;
    IF to_regclass('economy_tag') IS NOT NULL THEN
        INSERT INTO base_tag (id, key_type, key, name, created_at, updated_at)
        SELECT id, 'general', key, name, created_at, updated_at FROM economy_tag
        ON CONFLICT DO NOTHING;
    END IF;
END $$;
-- +migrate StatementEnd

ALTER TABLE economy_prize DROP CONSTRAINT IF EXISTS economy_prize_category_id_fkey;
ALTER TABLE economy_prize DROP CONSTRAINT IF EXISTS economy_prize_base_category_fkey;
ALTER TABLE economy_prize ADD CONSTRAINT economy_prize_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);

ALTER TABLE economy_product DROP CONSTRAINT IF EXISTS economy_product_category_id_fkey;
ALTER TABLE economy_product DROP CONSTRAINT IF EXISTS economy_product_base_category_fkey;
ALTER TABLE economy_product ADD CONSTRAINT economy_product_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);

ALTER TABLE publication_faq DROP CONSTRAINT IF EXISTS publication_faq_category_id_fkey;
ALTER TABLE publication_faq DROP CONSTRAINT IF EXISTS publication_faq_base_category_fkey;
ALTER TABLE publication_faq ADD CONSTRAINT publication_faq_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);

ALTER TABLE publication_news DROP CONSTRAINT IF EXISTS publication_news_category_id_fkey;
ALTER TABLE publication_news DROP CONSTRAINT IF EXISTS publication_news_base_category_fkey;
ALTER TABLE publication_news ADD CONSTRAINT publication_news_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);

DROP TABLE IF EXISTS economy_tag;
DROP TABLE IF EXISTS economy_category;

-- +migrate Down
-- +migrate StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Base migration % moved economy categories and tags into base_category and base_tag and dropped the old tables; restore from backup', '20261101000001-base-move-categories'; END $$;
-- +migrate StatementEnd
