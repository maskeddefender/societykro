-- ==========================================
-- Migration 000003: Complaint Tables
-- ==========================================

-- VENDOR (needed before complaint FK)
CREATE TABLE vendor (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  name                  VARCHAR(255) NOT NULL,
  company_name          VARCHAR(255),
  phone                 VARCHAR(15) NOT NULL,
  whatsapp_phone        VARCHAR(15),
  category              VARCHAR(50) NOT NULL,
  sub_category          VARCHAR(50),
  address               TEXT,
  avg_rating            DECIMAL(2,1) NOT NULL DEFAULT 0.0,
  total_jobs            INTEGER NOT NULL DEFAULT 0,
  completed_jobs        INTEGER NOT NULL DEFAULT 0,
  response_time_avg_hrs DECIMAL(5,1) DEFAULT 0.0,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vendor_society_category ON vendor(society_id, category);

-- COMPLAINT
CREATE TABLE complaint (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id              UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  flat_id                 UUID REFERENCES flat(id) ON DELETE SET NULL,
  raised_by               UUID NOT NULL REFERENCES app_user(id) ON DELETE RESTRICT,
  ticket_number           VARCHAR(20) NOT NULL UNIQUE,
  category                complaint_category NOT NULL,
  title                   VARCHAR(255) NOT NULL,
  description             TEXT NOT NULL,
  description_original    TEXT,
  description_english     TEXT,
  original_language       VARCHAR(10),
  voice_url               VARCHAR(500),
  voice_transcription     TEXT,
  image_urls              JSONB DEFAULT '[]',
  status                  complaint_status NOT NULL DEFAULT 'open',
  priority                complaint_priority NOT NULL DEFAULT 'normal',
  is_emergency            BOOLEAN NOT NULL DEFAULT false,
  is_common_area          BOOLEAN NOT NULL DEFAULT false,
  assigned_vendor_id      UUID REFERENCES vendor(id) ON DELETE SET NULL,
  assigned_by             UUID REFERENCES app_user(id),
  assigned_at             TIMESTAMPTZ,
  resolved_at             TIMESTAMPTZ,
  closed_at               TIMESTAMPTZ,
  reopened_at             TIMESTAMPTZ,
  resolution_rating       SMALLINT CHECK (resolution_rating BETWEEN 1 AND 5),
  resolution_feedback     TEXT,
  resolution_proof_urls   JSONB DEFAULT '[]',
  source                  message_source NOT NULL DEFAULT 'app',
  escalation_count        SMALLINT NOT NULL DEFAULT 0,
  escalation_deadline     TIMESTAMPTZ,
  is_recurring            BOOLEAN NOT NULL DEFAULT false,
  metadata                JSONB DEFAULT '{}',
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_complaint_society_status ON complaint(society_id, status);
CREATE INDEX idx_complaint_society_created ON complaint(society_id, created_at DESC);
CREATE INDEX idx_complaint_society_category ON complaint(society_id, category);
CREATE INDEX idx_complaint_flat ON complaint(flat_id) WHERE flat_id IS NOT NULL;
CREATE INDEX idx_complaint_vendor ON complaint(assigned_vendor_id) WHERE assigned_vendor_id IS NOT NULL;
CREATE INDEX idx_complaint_emergency ON complaint(society_id) WHERE is_emergency = true AND status != 'closed';
CREATE INDEX idx_complaint_escalation ON complaint(escalation_deadline) WHERE status = 'open';

-- COMPLAINT COMMENT
CREATE TABLE complaint_comment (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  complaint_id          UUID NOT NULL REFERENCES complaint(id) ON DELETE CASCADE,
  user_id               UUID NOT NULL REFERENCES app_user(id) ON DELETE RESTRICT,
  comment               TEXT NOT NULL,
  image_url             VARCHAR(500),
  is_internal           BOOLEAN NOT NULL DEFAULT false,
  is_status_change      BOOLEAN NOT NULL DEFAULT false,
  old_status            complaint_status,
  new_status            complaint_status,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cc_complaint ON complaint_comment(complaint_id, created_at);
