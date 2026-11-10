-- +migrate Up
ALTER TABLE economy_gift ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE economy_gift ADD COLUMN IF NOT EXISTS tag_ids TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE economy_gift DROP CONSTRAINT IF EXISTS economy_gift_base_category_fkey;
ALTER TABLE economy_gift ADD CONSTRAINT economy_gift_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);
CREATE INDEX IF NOT EXISTS economy_gift_category_idx ON economy_gift (category_id);

ALTER TABLE economy_currency ADD COLUMN IF NOT EXISTS category_id UUID;
ALTER TABLE economy_currency ADD COLUMN IF NOT EXISTS tag_ids TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE economy_currency DROP CONSTRAINT IF EXISTS economy_currency_base_category_fkey;
ALTER TABLE economy_currency ADD CONSTRAINT economy_currency_base_category_fkey FOREIGN KEY (category_id) REFERENCES base_category (id);
CREATE INDEX IF NOT EXISTS economy_currency_category_idx ON economy_currency (category_id);

ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','gift','currency','news','faq','general','user'));

-- +migrate Down
ALTER TABLE economy_gift DROP CONSTRAINT IF EXISTS economy_gift_base_category_fkey;
ALTER TABLE economy_gift DROP COLUMN IF EXISTS category_id;
ALTER TABLE economy_gift DROP COLUMN IF EXISTS tag_ids;
ALTER TABLE economy_currency DROP CONSTRAINT IF EXISTS economy_currency_base_category_fkey;
ALTER TABLE economy_currency DROP COLUMN IF EXISTS category_id;
ALTER TABLE economy_currency DROP COLUMN IF EXISTS tag_ids;
DELETE FROM base_tag WHERE key_type IN ('gift','currency');
ALTER TABLE base_tag DROP CONSTRAINT IF EXISTS base_tag_key_type_check;
ALTER TABLE base_tag ADD CONSTRAINT base_tag_key_type_check CHECK (key_type IN ('product','prize','news','faq','general','user'));
