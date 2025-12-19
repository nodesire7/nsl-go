-- 0004_rbac_permissions.sql
-- RBAC 权限点系统

-- 权限点表（预定义权限点）
CREATE TABLE IF NOT EXISTS permissions (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  resource_type VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户权限关联表（用户拥有的权限点）
CREATE TABLE IF NOT EXISTS user_permissions (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_permission_id ON user_permissions(permission_id);

-- 角色权限关联表（角色默认权限）
CREATE TABLE IF NOT EXISTS role_permissions (
  id SERIAL PRIMARY KEY,
  role VARCHAR(20) NOT NULL,
  permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(role, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);

-- 插入预定义权限点
INSERT INTO permissions (name, description, resource_type) VALUES
  ('link:create', '创建短链接', 'link'),
  ('link:delete', '删除短链接', 'link'),
  ('link:view', '查看短链接', 'link'),
  ('link:list', '列出短链接', 'link'),
  ('domain:manage', '管理域名', 'domain'),
  ('domain:create', '创建域名', 'domain'),
  ('domain:delete', '删除域名', 'domain'),
  ('settings:update', '更新系统设置', 'settings'),
  ('settings:view', '查看系统设置', 'settings'),
  ('user:manage', '管理用户', 'user'),
  ('user:view', '查看用户', 'user'),
  ('stats:view', '查看统计信息', 'stats')
ON CONFLICT (name) DO NOTHING;

-- admin 角色默认拥有所有权限
INSERT INTO role_permissions (role, permission_id)
SELECT 'admin', id FROM permissions
ON CONFLICT (role, permission_id) DO NOTHING;

-- user 角色默认权限（基础权限）
INSERT INTO role_permissions (role, permission_id)
SELECT 'user', id FROM permissions WHERE name IN (
  'link:create', 'link:delete', 'link:view', 'link:list',
  'domain:create', 'domain:delete', 'stats:view'
)
ON CONFLICT (role, permission_id) DO NOTHING;

