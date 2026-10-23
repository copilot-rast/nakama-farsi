-- Administrators of brands whose contract is active. A suspended or
-- terminated contract deactivates the brand, and the source tenant guard then
-- takes every brand-panel route down; authorization reads this view so an
-- administrator of an inactive brand holds no brand authority.
-- +migrate Up
CREATE OR REPLACE VIEW jazebe_active_brand_admin AS
    SELECT a.* FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id = a.brand_id WHERE b.is_active;

-- +migrate Down
-- A view holds no data.
DROP VIEW IF EXISTS jazebe_active_brand_admin;
