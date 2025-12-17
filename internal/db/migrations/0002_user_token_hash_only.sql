-- 0002_user_token_hash_only.sql
-- 安全基线：用户 API Token 不再明文存储（仅保留 hash）

-- 允许 api_token 为 NULL（避免空字符串冲突）
ALTER TABLE users ALTER COLUMN api_token DROP NOT NULL;

-- 移除 api_token 的唯一约束（后续不再写入明文）
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_api_token_key;

-- 移除明文 token 索引（不再需要）
DROP INDEX IF EXISTS idx_users_api_token;


