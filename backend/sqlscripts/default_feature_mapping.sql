INSERT INTO feature_menu_item (feature_id, menu_item_id)
SELECT 1, 1
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 1 AND menu_item_id = 1)
UNION ALL
SELECT 3, 3
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 3 AND menu_item_id = 3)
UNION ALL
SELECT 4, 4
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 4 AND menu_item_id = 4)
UNION ALL
SELECT 6, 6
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 6 AND menu_item_id = 6)
UNION ALL
SELECT 7, 7
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 7 AND menu_item_id = 7)
UNION ALL
SELECT 8, 8
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 8 AND menu_item_id = 8)
UNION ALL
SELECT 9, 9
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 9 AND menu_item_id = 9)
UNION ALL
SELECT 10, 10
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 10 AND menu_item_id = 10)
UNION ALL
SELECT 11, 11
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 11 AND menu_item_id = 11)
UNION ALL
SELECT 12, 12
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 12 AND menu_item_id = 12)
UNION ALL
SELECT 13, 13
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 13 AND menu_item_id = 13)
UNION ALL
SELECT 14, 14
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 14 AND menu_item_id = 14)
UNION ALL
SELECT 15, 15
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 15 AND menu_item_id = 15)
UNION ALL
SELECT 16, 16
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 16 AND menu_item_id = 16)
UNION ALL
SELECT 17, 17
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 17 AND menu_item_id = 17)
UNION ALL
SELECT 18, 18
    WHERE NOT EXISTS (SELECT 1 FROM feature_menu_item WHERE feature_id = 18 AND menu_item_id = 18);
