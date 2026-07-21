UPDATE users
SET is_admin = 'FALSE',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'admin@polirediucen.onmicrosoft.com';