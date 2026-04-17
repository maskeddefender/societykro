-- ==========================================
-- Migration 000005: Payment & Expense Tables
-- ==========================================

CREATE TABLE payment (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id) ON DELETE RESTRICT,
  flat_id               UUID NOT NULL REFERENCES flat(id) ON DELETE RESTRICT,
  user_id               UUID REFERENCES app_user(id),
  invoice_number        VARCHAR(30) NOT NULL UNIQUE,
  bill_month            DATE NOT NULL,
  base_amount           DECIMAL(12,2) NOT NULL,
  late_fee              DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  discount              DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  total_due             DECIMAL(12,2) NOT NULL,
  amount_paid           DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  status                payment_status NOT NULL DEFAULT 'pending',
  payment_method        payment_method,
  gateway_order_id      VARCHAR(100),
  gateway_payment_id    VARCHAR(100),
  gateway_signature     VARCHAR(255),
  paid_at               TIMESTAMPTZ,
  receipt_url           VARCHAR(500),
  due_date              DATE NOT NULL,
  reminder_count        SMALLINT NOT NULL DEFAULT 0,
  last_reminder_at      TIMESTAMPTZ,
  notes                 TEXT,
  recorded_by           UUID REFERENCES app_user(id),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(society_id, flat_id, bill_month)
);

CREATE INDEX idx_payment_society_status ON payment(society_id, status);
CREATE INDEX idx_payment_flat ON payment(flat_id, bill_month DESC);
CREATE INDEX idx_payment_overdue ON payment(due_date) WHERE status IN ('pending', 'overdue');
CREATE INDEX idx_payment_gateway ON payment(gateway_order_id) WHERE gateway_order_id IS NOT NULL;

-- EXPENSE
CREATE TABLE expense (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  society_id            UUID NOT NULL REFERENCES society(id),
  recorded_by           UUID NOT NULL REFERENCES app_user(id),
  category              VARCHAR(50) NOT NULL,
  description           TEXT NOT NULL,
  amount                DECIMAL(12,2) NOT NULL,
  expense_date          DATE NOT NULL,
  receipt_url           VARCHAR(500),
  vendor_id             UUID REFERENCES vendor(id),
  is_recurring          BOOLEAN NOT NULL DEFAULT false,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expense_society ON expense(society_id, expense_date DESC);
