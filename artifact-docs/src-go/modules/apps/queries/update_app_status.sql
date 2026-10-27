-- O resultado de uma conferência do monitor.
-- $1: id · $2: status ('online' | 'offline')
UPDATE apps
SET
  status     = $2,
  checked_at = now()
WHERE id = $1::UUID;
