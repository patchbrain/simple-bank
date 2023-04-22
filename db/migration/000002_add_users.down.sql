-- 删除确保owner,currency唯一的约束
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";

-- 删除owner与users表的链接的外键
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";

-- 删除users表
DROP TABLE  IF EXISTS "users";