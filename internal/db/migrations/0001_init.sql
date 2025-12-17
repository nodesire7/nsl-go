-- 0001_init.sql
-- 重写版初始表结构（与 legacy 兼容，后续将逐步替换）

CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- users
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  api_token VARCHAR(255) UNIQUE NOT NULL,
  api_token_hash VARCHAR(64),
  role VARCHAR(20) DEFAULT 'user',
  max_links INTEGER DEFAULT 10,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_token ON users(api_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_api_token_hash_unique ON users(api_token_hash) WHERE api_token_hash IS NOT NULL;

-- domains
CREATE TABLE IF NOT EXISTS domains (
  id SERIAL PRIMARY KEY,
  user_id BIGINT DEFAULT 0,
  domain VARCHAR(255) NOT NULL,
  is_default BOOLEAN DEFAULT false,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, domain)
);
CREATE INDEX IF NOT EXISTS idx_domains_user_id ON domains(user_id);
CREATE INDEX IF NOT EXISTS idx_domains_domain ON domains(domain);

-- links
CREATE TABLE IF NOT EXISTS links (
  id SERIAL PRIMARY KEY,
  user_id BIGINT DEFAULT 0,
  domain_id BIGINT DEFAULT 0,
  code VARCHAR(255) NOT NULL,
  original_url TEXT NOT NULL,
  title VARCHAR(500),
  hash VARCHAR(64) NOT NULL,
  qr_code TEXT,
  click_count BIGINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain_id, code)
);
CREATE INDEX IF NOT EXISTS idx_links_code ON links(code);
CREATE INDEX IF NOT EXISTS idx_links_hash ON links(hash);
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id);
CREATE INDEX IF NOT EXISTS idx_links_domain_id ON links(domain_id);
CREATE INDEX IF NOT EXISTS idx_links_created_at ON links(created_at DESC);

-- access_logs
CREATE TABLE IF NOT EXISTS access_logs (
  id SERIAL PRIMARY KEY,
  link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  ip VARCHAR(45),
  user_agent TEXT,
  referer TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_access_logs_link_id ON access_logs(link_id);
CREATE INDEX IF NOT EXISTS idx_access_logs_created_at ON access_logs(created_at DESC);

-- settings
CREATE TABLE IF NOT EXISTS settings (
  key VARCHAR(255) PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- default domain (system)
INSERT INTO domains (user_id, domain, is_default, is_active)
SELECT 0, '', true, true
WHERE NOT EXISTS (SELECT 1 FROM domains WHERE user_id = 0 AND is_default = true);


