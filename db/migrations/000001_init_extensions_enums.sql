-- ==========================================
-- Migration 000001: Extensions & ENUM Types
-- ==========================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";  -- For amenity booking exclusion constraint

-- ENUM Types
CREATE TYPE user_role AS ENUM (
  'resident', 'admin', 'secretary', 'treasurer', 'president', 'guard', 'vendor'
);

CREATE TYPE complaint_status AS ENUM (
  'open', 'in_progress', 'resolved', 'closed', 'reopened'
);

CREATE TYPE complaint_priority AS ENUM (
  'low', 'normal', 'high', 'emergency'
);

CREATE TYPE complaint_category AS ENUM (
  'plumbing', 'electrical', 'lift', 'water', 'security', 'parking',
  'garbage', 'pest_control', 'drain', 'generator', 'intercom',
  'common_area', 'noise', 'structural', 'other'
);

CREATE TYPE visitor_status AS ENUM (
  'pending', 'approved', 'denied', 'checked_in', 'checked_out', 'expired'
);

CREATE TYPE visitor_purpose AS ENUM (
  'guest', 'delivery', 'cab', 'service', 'official', 'other'
);

CREATE TYPE payment_status AS ENUM (
  'pending', 'paid', 'partial', 'overdue', 'waived'
);

CREATE TYPE payment_method AS ENUM (
  'upi', 'netbanking', 'card', 'cheque', 'cash', 'wallet', 'neft'
);

CREATE TYPE notice_category AS ENUM (
  'general', 'maintenance', 'emergency', 'event', 'meeting', 'rule', 'financial'
);

CREATE TYPE subscription_plan AS ENUM ('free', 'basic', 'premium', 'enterprise');
CREATE TYPE flat_type AS ENUM ('apartment', 'villa', 'shop', 'office', 'penthouse');
CREATE TYPE occupancy_type AS ENUM ('owner', 'tenant', 'vacant');
CREATE TYPE message_source AS ENUM ('app', 'whatsapp', 'telegram', 'web', 'voice', 'guard_app');
CREATE TYPE notification_channel AS ENUM ('push', 'whatsapp', 'telegram', 'sms', 'email');
