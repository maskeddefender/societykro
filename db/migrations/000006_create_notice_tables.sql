-- ==========================================
-- Migration 000006: Notice, Poll, Document, Listing Tables
-- ==========================================

-- NOTICE
CREATE TABLE notice (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  created_by            UUID NOT NULL REFERENCES app_user(id),
  title                 VARCHAR(255) NOT NULL,
  body                  TEXT NOT NULL,
  category              notice_category NOT NULL DEFAULT 'general',
  is_pinned             BOOLEAN NOT NULL DEFAULT false,
  broadcast_whatsapp    BOOLEAN NOT NULL DEFAULT false,
  broadcast_telegram    BOOLEAN NOT NULL DEFAULT false,
  broadcast_sms         BOOLEAN NOT NULL DEFAULT false,
  attachment_urls       JSONB DEFAULT '[]',
  expires_at            TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notice_society ON notice(society_id, created_at DESC);
CREATE INDEX idx_notice_pinned ON notice(society_id) WHERE is_pinned = true;

-- NOTICE READ RECEIPT
CREATE TABLE notice_read_receipt (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  notice_id             UUID NOT NULL REFERENCES notice(id) ON DELETE CASCADE,
  user_id               UUID NOT NULL REFERENCES app_user(id),
  read_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  channel               message_source NOT NULL DEFAULT 'app',
  UNIQUE(notice_id, user_id)
);

-- POLL
CREATE TABLE poll (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  created_by            UUID NOT NULL REFERENCES app_user(id),
  question              TEXT NOT NULL,
  options               JSONB NOT NULL,
  is_anonymous          BOOLEAN NOT NULL DEFAULT true,
  is_multiple           BOOLEAN NOT NULL DEFAULT false,
  closes_at             TIMESTAMPTZ,
  is_closed             BOOLEAN NOT NULL DEFAULT false,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_poll_society ON poll(society_id, created_at DESC);

-- POLL VOTE
CREATE TABLE poll_vote (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  poll_id               UUID NOT NULL REFERENCES poll(id) ON DELETE CASCADE,
  user_id               UUID NOT NULL REFERENCES app_user(id),
  flat_id               UUID NOT NULL REFERENCES flat(id),
  option_index          SMALLINT NOT NULL,
  voted_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(poll_id, flat_id)
);

-- SOCIETY DOCUMENT
CREATE TABLE society_document (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  uploaded_by           UUID NOT NULL REFERENCES app_user(id),
  title                 VARCHAR(255) NOT NULL,
  category              VARCHAR(50) NOT NULL,
  file_url              VARCHAR(500) NOT NULL,
  file_size_bytes       INTEGER,
  file_type             VARCHAR(10),
  is_public             BOOLEAN NOT NULL DEFAULT false,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_doc_society ON society_document(society_id, category);

-- LISTING (Rent/Sale)
CREATE TABLE listing (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  flat_id               UUID NOT NULL REFERENCES flat(id),
  listed_by             UUID NOT NULL REFERENCES app_user(id),
  listing_type          VARCHAR(10) NOT NULL,
  price                 DECIMAL(14,2) NOT NULL,
  description           TEXT,
  image_urls            JSONB DEFAULT '[]',
  contact_phone         VARCHAR(15) NOT NULL,
  is_active             BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_listing_society ON listing(society_id) WHERE is_active = true;
