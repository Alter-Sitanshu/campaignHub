-- =========================
-- Indexes (drop first)
-- =========================
DROP INDEX IF EXISTS idx_submissions_campaign_status;
DROP INDEX IF EXISTS idx_brand_name;
DROP INDEX IF EXISTS idx_tx_from_to;

-- =========================
-- Submissions
-- =========================
DROP TABLE IF EXISTS submissions;

-- =========================
-- Users & Roles
-- =========================
DROP TABLE IF EXISTS platform_links;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

-- =========================
-- Brands & Campaigns
-- =========================
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS status;
DROP TABLE IF EXISTS brands;

-- =========================
-- Transactions & Accounts
-- =========================
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS tx_status;
DROP TABLE IF EXISTS accounts;

-- =========================
-- Support Tickets
-- =========================
DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS ticket_status;