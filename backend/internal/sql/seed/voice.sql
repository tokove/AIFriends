INSERT INTO voices(name, voice_id, created_at)
VALUES
('干净清爽男', 'longanshuo', NOW()),
('睿智轻熟男', 'longanzhi', NOW()),
('磁性低音男', 'longxiaocheng_v2', NOW()),
('温婉邻家女', 'longxing_v2', NOW()),
('甜美娇气女', 'longfeifei_v2', NOW()),
('温暖春风女', 'longyan_v2', NOW())
ON CONFLICT (voice_id) DO NOTHING;