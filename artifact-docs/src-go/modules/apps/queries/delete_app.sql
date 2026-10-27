-- $1: id
DELETE FROM apps
WHERE id = $1::UUID;
