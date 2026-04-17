-- ==========================================
-- Migration 000007: Domestic Help, Emergency, Amenity, SOS, Audit
-- ==========================================

-- DOMESTIC HELP
CREATE TABLE domestic_help (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  name                  VARCHAR(255) NOT NULL,
  phone                 VARCHAR(15),
  photo_url             VARCHAR(500),
  role                  VARCHAR(30) NOT NULL,
  id_proof_type         VARCHAR(20),
  id_proof_hash         VARCHAR(64),
  is_verified           BOOLEAN NOT NULL DEFAULT false,
  avg_rating            DECIMAL(2,1) NOT NULL DEFAULT 0.0,
  entry_method          VARCHAR(20) NOT NULL DEFAULT 'otp',
  is_active             BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dh_society ON domestic_help(society_id) WHERE is_active = true;

-- DOMESTIC HELP <-> FLAT mapping
CREATE TABLE domestic_help_flat (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  domestic_help_id      UUID NOT NULL REFERENCES domestic_help(id) ON DELETE CASCADE,
  flat_id               UUID NOT NULL REFERENCES flat(id),
  monthly_pay           DECIMAL(10,2),
  working_days          JSONB DEFAULT '["mon","tue","wed","thu","fri","sat"]',
  is_active             BOOLEAN NOT NULL DEFAULT true,
  UNIQUE(domestic_help_id, flat_id)
);

-- DOMESTIC HELP ATTENDANCE
CREATE TABLE domestic_help_attendance (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  domestic_help_id      UUID NOT NULL REFERENCES domestic_help(id) ON DELETE CASCADE,
  society_id            UUID NOT NULL REFERENCES society(id),
  flat_id               UUID NOT NULL REFERENCES flat(id),
  entry_at              TIMESTAMPTZ NOT NULL,
  exit_at               TIMESTAMPTZ,
  date                  DATE NOT NULL,
  source                message_source NOT NULL DEFAULT 'guard_app'
);

CREATE INDEX idx_dha_help_date ON domestic_help_attendance(domestic_help_id, date);
CREATE INDEX idx_dha_flat_date ON domestic_help_attendance(flat_id, date);

-- EMERGENCY CONTACT
CREATE TABLE emergency_contact (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  name                  VARCHAR(255) NOT NULL,
  category              VARCHAR(30) NOT NULL,
  phone                 VARCHAR(15) NOT NULL,
  address               TEXT,
  is_default            BOOLEAN NOT NULL DEFAULT false,
  sort_order            SMALLINT NOT NULL DEFAULT 0,
  is_active             BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_ec_society ON emergency_contact(society_id) WHERE is_active = true;

-- SOS ALERT
CREATE TABLE sos_alert (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  triggered_by          UUID NOT NULL REFERENCES app_user(id),
  flat_id               UUID REFERENCES flat(id),
  alert_type            VARCHAR(20) NOT NULL,
  message               TEXT,
  latitude              DECIMAL(10,8),
  longitude             DECIMAL(11,8),
  is_resolved           BOOLEAN NOT NULL DEFAULT false,
  resolved_by           UUID REFERENCES app_user(id),
  resolved_at           TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sos_society ON sos_alert(society_id, created_at DESC);

-- AMENITY
CREATE TABLE amenity (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  name                  VARCHAR(100) NOT NULL,
  description           TEXT,
  capacity              INTEGER,
  booking_fee           DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  deposit_fee           DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  requires_approval     BOOLEAN NOT NULL DEFAULT false,
  max_hours             SMALLINT DEFAULT 4,
  available_from        TIME DEFAULT '06:00',
  available_to          TIME DEFAULT '22:00',
  rules                 TEXT,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_amenity_society ON amenity(society_id) WHERE is_active = true;

-- AMENITY BOOKING
CREATE TABLE amenity_booking (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  amenity_id            UUID NOT NULL REFERENCES amenity(id),
  user_id               UUID NOT NULL REFERENCES app_user(id),
  flat_id               UUID NOT NULL REFERENCES flat(id),
  society_id            UUID NOT NULL REFERENCES society(id),
  booking_date          DATE NOT NULL,
  start_time            TIME NOT NULL,
  end_time              TIME NOT NULL,
  purpose               VARCHAR(255),
  guest_count           SMALLINT DEFAULT 0,
  status                VARCHAR(20) NOT NULL DEFAULT 'pending',
  approved_by           UUID REFERENCES app_user(id),
  cancelled_at          TIMESTAMPTZ,
  cancel_reason         VARCHAR(255),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_amenity_date ON amenity_booking(amenity_id, booking_date);

-- AUDIT LOG
CREATE TABLE audit_log (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL,
  user_id               UUID NOT NULL,
  action                VARCHAR(100) NOT NULL,
  entity_type           VARCHAR(50) NOT NULL,
  entity_id             UUID NOT NULL,
  old_value             JSONB,
  new_value             JSONB,
  ip_address            INET,
  user_agent            TEXT,
  source                message_source NOT NULL DEFAULT 'app',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_society ON audit_log(society_id, created_at DESC);
CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);
