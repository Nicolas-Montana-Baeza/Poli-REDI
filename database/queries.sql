BEGIN TRAN;

SELECT id, name, description, is_active
FROM dbo.activities
WHERE name LIKE N'%---%';

DELETE FROM dbo.activities
WHERE name LIKE N'%---%';

-- Revisa que ya no aparezca
SELECT id, name
FROM dbo.activities
WHERE name LIKE N'%---%';

COMMIT;
SELECT * FROM activities;