-- === DEBUG SCRIPT PARA NOTIFICATIONS ===
-- Reemplaza los UUIDs con tus valores reales

SET @STREAMER_ID = 'tu-uuid-del-streamer';
SET @FOLLOWER_ID = 'tu-uuid-del-follower';

-- 1. Verificar si existe la relación de followers
SELECT 'FOLLOWERS RELATIONSHIP' as check_name;
SELECT f.* FROM followers f 
WHERE f.streamer_id = @STREAMER_ID AND f.follower_id = @FOLLOWER_ID;

-- 2. Contar seguidores del streamer
SELECT 'FOLLOWERS COUNT' as check_name, COUNT(*) as total FROM followers 
WHERE streamer_id = @STREAMER_ID;

-- 3. Ver todos los tokens del follower
SELECT 'FOLLOWER TOKENS' as check_name;
SELECT dt.id, dt.user_id, dt.token, dt.platform, dt.is_valid, dt.created_at, dt.updated_at 
FROM device_tokens dt 
WHERE dt.user_id = @FOLLOWER_ID;

-- 4. Ver tokens válidos del follower
SELECT 'FOLLOWER VALID TOKENS' as check_name;
SELECT dt.id, dt.user_id, dt.token, dt.platform, dt.created_at 
FROM device_tokens dt 
WHERE dt.user_id = @FOLLOWER_ID AND dt.is_valid = true;

-- 5. Simular la query que usa NotifyStreamLive
SELECT 'QUERY SIMULATION' as check_name;
SELECT dt.id, dt.user_id, dt.token, dt.platform, dt.device_id, dt.app_version,
       dt.is_valid, dt.last_used_at, dt.created_at, dt.updated_at
FROM device_tokens dt
INNER JOIN followers f ON f.follower_id = dt.user_id
WHERE f.streamer_id = @STREAMER_ID AND dt.is_valid = true;

-- 6. Ver tokens inválidos que podrían ser bloqueados
SELECT 'INVALID TOKENS' as check_name, COUNT(*) as total 
FROM device_tokens dt 
WHERE dt.is_valid = false;

-- 7. Ver todos los tokens del streamer
SELECT 'STREAMER TOKENS' as check_name;
SELECT dt.id, dt.user_id, dt.token, dt.platform, dt.is_valid, dt.created_at 
FROM device_tokens dt 
WHERE dt.user_id = @STREAMER_ID;
