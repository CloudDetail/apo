INSERT INTO auth_permission (subject_id, subject_type, type, permission_id)
SELECT 1, 'role', 'feature', 1 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 1)
UNION ALL
SELECT 1, 'role', 'feature', 2 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 2)
UNION ALL
SELECT 1, 'role', 'feature', 3 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 3)
UNION ALL
SELECT 1, 'role', 'feature', 4 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 4)
UNION ALL
SELECT 1, 'role', 'feature', 5 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 5)
UNION ALL
SELECT 1, 'role', 'feature', 6 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 6)
UNION ALL
SELECT 1, 'role', 'feature', 7 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 7)
UNION ALL
SELECT 1, 'role', 'feature', 8 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 8)
UNION ALL
SELECT 1, 'role', 'feature', 9 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 9)
UNION ALL
SELECT 1, 'role', 'feature', 11 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 11)
UNION ALL
SELECT 1, 'role', 'feature', 12 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 12)
UNION ALL
SELECT 1, 'role', 'feature', 14 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 14)
UNION ALL
SELECT 1, 'role', 'feature', 15 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 15)
UNION ALL
SELECT 1, 'role', 'feature', 16 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 16)
UNION ALL
SELECT 1, 'role', 'feature', 17 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 17)
UNION ALL
SELECT 1, 'role', 'feature', 18 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 1 AND subject_type = 'role' AND type = 'feature' AND permission_id = 18)
UNION ALL
SELECT 2, 'role', 'feature', 1 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 1)
UNION ALL
SELECT 2, 'role', 'feature', 2 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 2)
UNION ALL
SELECT 2, 'role', 'feature', 3 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 3)
UNION ALL
SELECT 2, 'role', 'feature', 4 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 4)
UNION ALL
SELECT 2, 'role', 'feature', 5 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 5)
UNION ALL
SELECT 2, 'role', 'feature', 6 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 6)
UNION ALL
SELECT 2, 'role', 'feature', 7 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 7)
UNION ALL
SELECT 2, 'role', 'feature', 8 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 8)
UNION ALL
SELECT 2, 'role', 'feature', 9 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 9)
UNION ALL
SELECT 2, 'role', 'feature', 11 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 11)
UNION ALL
SELECT 2, 'role', 'feature', 12 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 12)
UNION ALL
SELECT 2, 'role', 'feature', 17 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 17)
UNION ALL
SELECT 2, 'role', 'feature', 18 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 2 AND subject_type = 'role' AND type = 'feature' AND permission_id = 18)
UNION ALL
SELECT 3, 'role', 'feature', 1 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 1)
UNION ALL
SELECT 3, 'role', 'feature', 2 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 2)
UNION ALL
SELECT 3, 'role', 'feature', 3 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 3)
UNION ALL
SELECT 3, 'role', 'feature', 4 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 4)
UNION ALL
SELECT 3, 'role', 'feature', 5 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 5)
UNION ALL
SELECT 3, 'role', 'feature', 6 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 6)
UNION ALL
SELECT 3, 'role', 'feature', 7 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 7)
UNION ALL
SELECT 3, 'role', 'feature', 8 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 8)
UNION ALL
SELECT 3, 'role', 'feature', 9 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 3 AND subject_type = 'role' AND type = 'feature' AND permission_id = 9)
UNION ALL
SELECT 4, 'role', 'feature', 1 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 1)
UNION ALL
SELECT 4, 'role', 'feature', 2 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 2)
UNION ALL
SELECT 4, 'role', 'feature', 3 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 3)
UNION ALL
SELECT 4, 'role', 'feature', 4 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 4)
UNION ALL
SELECT 4, 'role', 'feature', 5 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 5)
UNION ALL
SELECT 4, 'role', 'feature', 6 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 6)
UNION ALL
SELECT 4, 'role', 'feature', 7 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 7)
UNION ALL
SELECT 4, 'role', 'feature', 8 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 8)
UNION ALL
SELECT 4, 'role', 'feature', 9 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 9)
UNION ALL
SELECT 4, 'role', 'feature', 10 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 10)
UNION ALL
SELECT 4, 'role', 'feature', 11 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 11)
UNION ALL
SELECT 4, 'role', 'feature', 13 WHERE NOT EXISTS (SELECT 1 FROM auth_permission WHERE subject_id = 4 AND subject_type = 'role' AND type = 'feature' AND permission_id = 13)
