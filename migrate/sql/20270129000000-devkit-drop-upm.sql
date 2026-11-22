-- +migrate Up
DELETE FROM devkit_api_scope WHERE source = 'devkit' AND name IN ('upm:read', 'upm:publish');
UPDATE devkit_api_key SET scopes = array_remove(array_remove(scopes, 'upm:read'), 'upm:publish') WHERE scopes && ARRAY['upm:read', 'upm:publish']::TEXT[];
UPDATE devkit_settings SET value = value - 'packages' WHERE value ? 'packages';
DELETE FROM devkit_error_message WHERE code LIKE 'UPM\_%' AND NOT is_custom;

-- +migrate Down
SELECT 1;
