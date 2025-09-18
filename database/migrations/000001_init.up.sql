CREATE EXTENSION IF NOT EXISTS citext;

-- =========================
-- Support Tickets
-- =========================
CREATE TABLE support_tickets (
  id varchar(36) PRIMARY KEY,
  customer_id varchar(36) NOT NULL,  -- could be user_id or brand_id (if polymorphic, enforce in app/trigger)
  subject varchar(100) NOT NULL,
  message text NOT NULL,
  status int NOT NULL,
  created_at timestamptz DEFAULT now(),
  type VARCHAR(10) NOT NULL CHECK (type IN ('brand', 'creator'))
);

CREATE TABLE ticket_status (
  id int PRIMARY KEY,
  name varchar(10) NOT NULL CHECK (name IN ('open', 'resolved'))
);

-- =========================
-- Transactions & Accounts
-- =========================
CREATE TABLE accounts (
  id varchar(36) PRIMARY KEY,
  holder_id varchar(36) NOT NULL,
  holder_type varchar(10) NOT NULL CHECK (holder_type IN ('user','brand')),
  amount numeric(12,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
  active boolean DEFAULT TRUE,
  created_at timestamptz DEFAULT now(),
  UNIQUE (holder_id, holder_type)
);

CREATE TABLE tx_status (
  id int PRIMARY KEY,
  name varchar(10) NOT NULL CHECK (name IN ('success','failed'))
);

CREATE TABLE transactions (
  id varchar(36) PRIMARY KEY,
  from_id varchar(36) NOT NULL,
  to_id varchar(36) NOT NULL,
  amount numeric(12,2) NOT NULL CHECK (amount > 0),
  status int NOT NULL,
  type varchar(20) NOT NULL CHECK (type IN ('withdraw', 'payout', 'deposit')),
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_tx_from FOREIGN KEY (from_id) REFERENCES accounts (id) ,
  CONSTRAINT fk_tx_to FOREIGN KEY (to_id) REFERENCES accounts (id),
  CONSTRAINT fk_tx_status FOREIGN KEY (status) REFERENCES tx_status (id)
);

-- =========================
-- Brands & Campaigns
-- =========================
CREATE TABLE brands (
  id varchar(36) PRIMARY KEY,
  name varchar(100) NOT NULL,
  email citext UNIQUE NOT NULL,
  sector varchar(20) NOT NULL,
  password BYTEA NOT NULL,
  website varchar(255) NOT NULL,
  address text NOT NULL,
  campaigns int NOT NULL DEFAULT 0
);

CREATE TABLE status (
  id int PRIMARY KEY,
  name varchar(10) NOT NULL CHECK (name IN ('active','draft','expired'))
);

CREATE TABLE campaigns (
  id varchar(36) PRIMARY KEY,
  brand_id varchar(36) NOT NULL,
  title varchar(100) NOT NULL,
  budget numeric(12,2) NOT NULL CHECK (budget >= 1000),
  cpm float NOT NULL CHECK (cpm >= 10.0),
  requirements text,
  platform varchar(20) NOT NULL,
  doc_link varchar(255),
  status int NOT NULL,
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_campaign_status FOREIGN KEY (status) REFERENCES status (id),
  CONSTRAINT fk_campaign_brand FOREIGN KEY (brand_id) REFERENCES brands (id)
);

-- =========================
-- Users & Roles
-- =========================
CREATE TABLE roles (
  id varchar(5) PRIMARY KEY,
  name varchar(10) NOT NULL
);

CREATE TABLE users (
  id varchar(36) PRIMARY KEY,
  first_name varchar(50) NOT NULL,
  last_name varchar(50),
  email citext UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  gender varchar(1) CHECK (gender IN ('M','F','O')),
  age int CHECK (age > 0 AND age < 100),
  role varchar(5) NOT NULL,
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_user_role FOREIGN KEY (role) REFERENCES roles (id)
);

CREATE TABLE platform_links (
  userid varchar(36) NOT NULL,
  platform varchar(10) NOT NULL,
  url varchar(255) NOT NULL,
  PRIMARY KEY (userid, platform),
  CONSTRAINT fk_platform_user FOREIGN KEY (userid) REFERENCES users (id)
);

-- =========================
-- Submissions
-- =========================
CREATE TABLE submissions (
  id varchar(36) PRIMARY KEY,
  creator_id varchar(36) NOT NULL,
  campaign_id varchar(36) NOT NULL,
  url varchar(255) NOT NULL,
  status int NOT NULL,
  views int NOT NULL DEFAULT 0 CHECK (views >= 0),
  earnings numeric(12,2) NOT NULL DEFAULT 0 CHECK (earnings >= 0),
  created_at timestamptz DEFAULT now(),
  CONSTRAINT fk_submission_user FOREIGN KEY (creator_id) REFERENCES users (id),
  CONSTRAINT fk_submission_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns (id),
  CONSTRAINT fk_submission_status FOREIGN KEY (status) REFERENCES status (id)
);

-- =========================
-- Indexes
-- =========================
CREATE INDEX idx_tx_from_to ON transactions (from_id, to_id);
CREATE INDEX idx_brand_name ON brands (name);
CREATE INDEX idx_submissions_campaign_status ON submissions (campaign_id, status);


INSERT INTO status (id, name)
VALUES (0, 'draft'), (1, 'active'), (3, 'expired');

INSERT INTO tx_status (id, name)
VALUES (0, 'failed'), (1, 'success');

INSERT INTO roles (id, name)
VALUES ('sup', 'superuser'), ('LVL1', 'base');

INSERT INTO ticket_status (id, name)
VALUES (0, 'resolved'), (1, 'open');