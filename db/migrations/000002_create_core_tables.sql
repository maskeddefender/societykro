-- ==========================================
-- Migration 000002: Core Tables (Society, Flat, User, Membership)
-- ==========================================

-- SOCIETY
CREATE TABLE society (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                  VARCHAR(255) NOT NULL,
  code                  VARCHAR(10) NOT NULL UNIQUE,
  address               TEXT NOT NULL,
  city                  VARCHAR(100) NOT NULL,
  state                 VARCHAR(100) NOT NULL,
  pincode               VARCHAR(6) NOT NULL,
  country               VARCHAR(50) NOT NULL DEFAULT 'India',
  total_flats           INTEGER NOT NULL CHECK (total_flats > 0),
  total_blocks          INTEGER DEFAULT 1,
  total_floors          INTEGER DEFAULT 1,
  subscription          subscription_plan NOT NULL DEFAULT 'free',
  subscription_expires  TIMESTAMPTZ,
  timezone              VARCHAR(50) NOT NULL DEFAULT 'Asia/Kolkata',
  default_language      VARCHAR(10) NOT NULL DEFAULT 'hi',
  maintenance_amount    DECIMAL(10,2) DEFAULT 0.00,
  maintenance_due_day   SMALLINT DEFAULT 1 CHECK (maintenance_due_day BETWEEN 1 AND 28),
  late_fee_percent      DECIMAL(4,2) DEFAULT 0.00,
  late_fee_grace_days   SMALLINT DEFAULT 15,
  escalation_hours      JSONB DEFAULT '{"water": 4, "lift": 8, "electrical": 12, "default": 24}',
  settings              JSONB DEFAULT '{}',
  logo_url              VARCHAR(500),
  is_active             BOOLEAN NOT NULL DEFAULT true,
  deleted_at            TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_society_code ON society(code);
CREATE INDEX idx_society_city ON society(city);
CREATE INDEX idx_society_pincode ON society(pincode);
CREATE INDEX idx_society_active ON society(is_active) WHERE is_active = true;

-- FLAT
CREATE TABLE flat (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  flat_number           VARCHAR(20) NOT NULL,
  block                 VARCHAR(50),
  floor                 SMALLINT,
  flat_type             flat_type NOT NULL DEFAULT 'apartment',
  area_sqft             INTEGER,
  bedrooms              SMALLINT,
  is_occupied           BOOLEAN NOT NULL DEFAULT false,
  occupancy             occupancy_type DEFAULT 'vacant',
  parking_slots         JSONB DEFAULT '[]',
  intercom_number       VARCHAR(20),
  maintenance_override  DECIMAL(10,2),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(society_id, flat_number)
);

CREATE INDEX idx_flat_society ON flat(society_id);
CREATE INDEX idx_flat_society_block ON flat(society_id, block);

-- USER
CREATE TABLE app_user (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone                 VARCHAR(15) NOT NULL UNIQUE,
  phone_hash            VARCHAR(64) NOT NULL,
  name                  VARCHAR(255) NOT NULL,
  email                 VARCHAR(255),
  avatar_url            VARCHAR(500),
  preferred_language    VARCHAR(10) NOT NULL DEFAULT 'hi',
  is_senior_citizen     BOOLEAN NOT NULL DEFAULT false,
  whatsapp_opted_in     BOOLEAN NOT NULL DEFAULT false,
  telegram_chat_id      VARCHAR(50),
  fcm_token             TEXT,
  apns_token            TEXT,
  device_info           JSONB DEFAULT '{}',
  last_active_at        TIMESTAMPTZ,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  deleted_at            TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_phone ON app_user(phone);
CREATE INDEX idx_user_phone_hash ON app_user(phone_hash);
CREATE INDEX idx_user_telegram ON app_user(telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;

-- USER-SOCIETY MEMBERSHIP
CREATE TABLE user_society_membership (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               UUID NOT NULL REFERENCES app_user(id) ON DELETE RESTRICT,
  society_id            UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  flat_id               UUID REFERENCES flat(id) ON DELETE SET NULL,
  role                  user_role NOT NULL DEFAULT 'resident',
  is_primary_member     BOOLEAN NOT NULL DEFAULT false,
  joined_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  left_at               TIMESTAMPTZ,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  UNIQUE(user_id, society_id)
);

CREATE INDEX idx_usm_society ON user_society_membership(society_id);
CREATE INDEX idx_usm_user ON user_society_membership(user_id);
CREATE INDEX idx_usm_flat ON user_society_membership(flat_id) WHERE flat_id IS NOT NULL;
CREATE INDEX idx_usm_society_role ON user_society_membership(society_id, role);
