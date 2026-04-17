-- ==========================================
-- Seed: Test data for local development
-- ==========================================

-- Test Society
INSERT INTO society (id, name, code, address, city, state, pincode, total_flats, total_blocks, total_floors, maintenance_amount, maintenance_due_day)
VALUES (
  'a0000000-0000-0000-0000-000000000001',
  'Green Valley Apartments',
  'GVA123',
  '123, MG Road, Koregaon Park',
  'Pune',
  'Maharashtra',
  '411001',
  24,
  2,
  6,
  3500.00,
  1
) ON CONFLICT DO NOTHING;

-- Flats: Block A (A-101 to A-106, A-201 to A-206, ... A-601 to A-606 = but let's do 12 flats per block)
INSERT INTO flat (society_id, flat_number, block, floor, flat_type, bedrooms) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'A-101', 'Block A', 1, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'A-102', 'Block A', 1, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'A-201', 'Block A', 2, 'apartment', 3),
  ('a0000000-0000-0000-0000-000000000001', 'A-202', 'Block A', 2, 'apartment', 3),
  ('a0000000-0000-0000-0000-000000000001', 'A-301', 'Block A', 3, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'A-302', 'Block A', 3, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'A-401', 'Block A', 4, 'apartment', 3),
  ('a0000000-0000-0000-0000-000000000001', 'A-402', 'Block A', 4, 'apartment', 3),
  ('a0000000-0000-0000-0000-000000000001', 'B-101', 'Block B', 1, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'B-102', 'Block B', 1, 'apartment', 2),
  ('a0000000-0000-0000-0000-000000000001', 'B-201', 'Block B', 2, 'apartment', 3),
  ('a0000000-0000-0000-0000-000000000001', 'B-202', 'Block B', 2, 'apartment', 3)
ON CONFLICT DO NOTHING;

-- Test Users
INSERT INTO app_user (id, phone, phone_hash, name, preferred_language) VALUES
  ('b0000000-0000-0000-0000-000000000001', '+919876543210', encode(sha256('+919876543210'::bytea), 'hex'), 'Sharma Ji', 'hi'),
  ('b0000000-0000-0000-0000-000000000002', '+919876543211', encode(sha256('+919876543211'::bytea), 'hex'), 'Priya Desai', 'en'),
  ('b0000000-0000-0000-0000-000000000003', '+919876543212', encode(sha256('+919876543212'::bytea), 'hex'), 'Mehta Bhai', 'hi'),
  ('b0000000-0000-0000-0000-000000000004', '+919876543213', encode(sha256('+919876543213'::bytea), 'hex'), 'Ramesh (Guard)', 'hi')
ON CONFLICT DO NOTHING;

-- Memberships
INSERT INTO user_society_membership (user_id, society_id, flat_id, role, is_primary_member) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
   (SELECT id FROM flat WHERE flat_number = 'A-301' AND society_id = 'a0000000-0000-0000-0000-000000000001'), 'resident', true),
  ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
   (SELECT id FROM flat WHERE flat_number = 'B-101' AND society_id = 'a0000000-0000-0000-0000-000000000001'), 'resident', true),
  ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
   (SELECT id FROM flat WHERE flat_number = 'A-101' AND society_id = 'a0000000-0000-0000-0000-000000000001'), 'secretary', true),
  ('b0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
   NULL, 'guard', false)
ON CONFLICT DO NOTHING;

-- Mark occupied flats
UPDATE flat SET is_occupied = true, occupancy = 'owner'
WHERE flat_number IN ('A-301', 'B-101', 'A-101')
  AND society_id = 'a0000000-0000-0000-0000-000000000001';

-- Test Vendors
INSERT INTO vendor (society_id, name, phone, category) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'Suresh Plumbing Services', '+919876500001', 'plumbing'),
  ('a0000000-0000-0000-0000-000000000001', 'Rajesh Electricals', '+919876500002', 'electrical'),
  ('a0000000-0000-0000-0000-000000000001', 'City Lift Services', '+919876500003', 'lift')
ON CONFLICT DO NOTHING;

-- Emergency Contacts
INSERT INTO emergency_contact (society_id, name, category, phone, is_default, sort_order) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'Sassoon Hospital', 'hospital', '02026127000', true, 1),
  ('a0000000-0000-0000-0000-000000000001', 'Pune Fire Brigade', 'fire', '101', true, 2),
  ('a0000000-0000-0000-0000-000000000001', 'Koregaon Park Police Station', 'police', '02026123456', true, 3),
  ('a0000000-0000-0000-0000-000000000001', 'Ambulance', 'ambulance', '108', true, 4)
ON CONFLICT DO NOTHING;
