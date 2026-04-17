-- ==========================================
-- Migration 000004: Visitor Tables
-- ==========================================

CREATE TABLE visitor (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  flat_id               UUID NOT NULL REFERENCES flat(id) ON DELETE RESTRICT,
  visitor_name          VARCHAR(255) NOT NULL,
  visitor_phone         VARCHAR(15),
  visitor_photo_url     VARCHAR(500),
  purpose               visitor_purpose NOT NULL DEFAULT 'guest',
  purpose_detail        VARCHAR(255),
  vehicle_number        VARCHAR(20),
  vehicle_type          VARCHAR(20),
  otp_code              VARCHAR(6),
  otp_expires_at        TIMESTAMPTZ,
  status                visitor_status NOT NULL DEFAULT 'pending',
  approved_by           UUID REFERENCES app_user(id),
  approved_via          message_source,
  denial_reason         VARCHAR(255),
  checked_in_at         TIMESTAMPTZ,
  checked_out_at        TIMESTAMPTZ,
  logged_by_guard       UUID REFERENCES app_user(id),
  is_pre_approved       BOOLEAN NOT NULL DEFAULT false,
  is_blacklisted        BOOLEAN NOT NULL DEFAULT false,
  source                message_source NOT NULL DEFAULT 'guard_app',
  metadata              JSONB DEFAULT '{}',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_visitor_society_date ON visitor(society_id, created_at DESC);
CREATE INDEX idx_visitor_flat ON visitor(flat_id, created_at DESC);
CREATE INDEX idx_visitor_otp ON visitor(society_id, otp_code) WHERE otp_code IS NOT NULL;
CREATE INDEX idx_visitor_active ON visitor(society_id) WHERE status IN ('pending', 'checked_in');

-- VISITOR PASS (recurring visitor auto-approval)
CREATE TABLE visitor_pass (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  flat_id               UUID NOT NULL REFERENCES flat(id),
  created_by            UUID NOT NULL REFERENCES app_user(id),
  visitor_name          VARCHAR(255) NOT NULL,
  visitor_phone         VARCHAR(15),
  pass_type             VARCHAR(20) NOT NULL DEFAULT 'recurring',
  valid_days            JSONB DEFAULT '["mon","tue","wed","thu","fri","sat"]',
  valid_from_time       TIME,
  valid_to_time         TIME,
  valid_from_date       DATE NOT NULL,
  valid_to_date         DATE,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vpass_society_flat ON visitor_pass(society_id, flat_id) WHERE is_active = true;
