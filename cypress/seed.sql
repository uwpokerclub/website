-- Seed the e2e testing user (webmaster)
INSERT INTO logins (username, password, role) VALUES ('e2e_user', '$2a$10$lzRaELvZxS2JwGsI0jSQueJWvMGfx82iYBuu0nFDCxuwJMabOHoX.', 'webmaster');

-- Seed additional logins for testing logins management
-- Password for all test logins is 'password123' (bcrypt hash)
INSERT INTO logins (username, password, role) VALUES
  ('test_president', '$2a$10$lzRaELvZxS2JwGsI0jSQueJWvMGfx82iYBuu0nFDCxuwJMabOHoX.', 'president'),
  ('test_executive', '$2a$10$lzRaELvZxS2JwGsI0jSQueJWvMGfx82iYBuu0nFDCxuwJMabOHoX.', 'executive'),
  ('hdrust0', '$2a$10$lzRaELvZxS2JwGsI0jSQueJWvMGfx82iYBuu0nFDCxuwJMabOHoX.', 'executive');

-- Seed a testing semester
INSERT INTO semesters 
  (id, name, start_date, end_date, starting_budget, current_budget, membership_fee, membership_discount_fee, rebuy_fee, meta) 
  VALUES ('84f026be-53e0-4759-ab89-131c4a66d649', 'Winter 2025', '2025-01-01', '2024-04-30', 100, 100, 10, 7, 2, 'Seed Semester');

-- Seed a initial event structure
INSERT INTO structures (id, name) VALUES (1, 'Structure A');
SELECT setval('structures_id_seq', (SELECT MAX(id) FROM structures));

-- Seed an initial set of blinds for the structure above
INSERT INTO blinds (id, small, big, ante, time, index, structure_id) VALUES
  (1, 10, 20, 20, 5, 0, 1),
  (2, 20, 40, 40, 5, 1, 1),
  (3, 30, 60, 60, 5, 2, 1),
  (4, 40, 80, 80, 5, 3, 1),
  (5, 50, 100, 100, 5, 4, 1);
SELECT setval('blinds_id_seq', (SELECT MAX(id) FROM blinds));

-- Seed events for the seed semester
INSERT INTO events
  (id, name, format, notes, semester_id, start_date, state, structure_id, rebuys, points_multiplier)
VALUES
  (1, 'Winter 2025 Event #1', 'No Limit Hold''em', 'Seed event', '84f026be-53e0-4759-ab89-131c4a66d649', '2025-01-03 19:00:00', 0, 1, 0, 1.0),
  (2, 'Winter 2025 Event #2', 'No Limit Hold''em', 'Completed event', '84f026be-53e0-4759-ab89-131c4a66d649', '2025-01-10 19:00:00', 1, 1, 0, 1.0);
SELECT setval('events_id_seq', (SELECT MAX(id) FROM events));

-- Seed users
INSERT INTO users (id, first_name, last_name, email, faculty, quest_id, created_at) VALUES
  (62958169, 'Heinrik', 'Drust', 'hdrust0@merriam-webster.com', 'AHS', 'hdrust0', '2025-04-18'),
  (20141158, 'Doretta', 'Housegoe', 'dhousegoe1@xinhuanet.com', 'AHS', 'dhousegoe1', '2025-04-19'),
  (85018940, 'Elita', 'Aucock', 'eaucock2@si.edu', 'Science', 'eaucock2', '2024-10-07'),
  (77679767, 'Khalil', 'Duckham', 'kduckham3@wsj.com', 'Engineering', 'kduckham3', '2025-03-22'),
  (70492884, 'Amandie', 'Libbis', 'alibbis4@google.co.jp', 'Engineering', 'alibbis4', '2025-03-31'),
  (39166759, 'Wald', 'Sundin', 'wsundin5@prweb.com', 'Science', 'wsundin5', '2024-06-30'),
  (55686346, 'Germayne', 'Croom', 'gcroom6@drupal.org', 'Science', 'gcroom6', '2024-05-28'),
  (81085720, 'Oralie', 'Bunten', 'obunten7@dropbox.com', 'Science', 'obunten7', '2024-07-19'),
  (52873146, 'Kristel', 'Callan', 'kcallan8@arizona.edu', 'Math', 'kcallan8', '2024-06-27'),
  (75969632, 'Winslow', 'Josey', 'wjosey9@blogger.com', 'Environment', 'wjosey9', '2025-01-30');

-- Seed users WITHOUT memberships (for testing registration flows)
INSERT INTO users (id, first_name, last_name, email, faculty, quest_id, created_at) VALUES
  (11111111, 'Unregistered', 'TestUser', 'unregistered@test.com', 'Math', 'unreg1', '2025-01-01'),
  (22222222, 'Another', 'Unregistered', 'another.unreg@test.com', 'Science', 'unreg2', '2025-01-01');

-- Seed members into the seed semester
INSERT INTO memberships (id, user_id, semester_id, paid, discounted) VALUES
  ('5d312426-ad56-4231-bb12-241acbfb91e2', 62958169, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d', 20141158, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e', 85018940, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('d4e5f6a7-b8c9-4d0e-a1b2-c3d4e5f6a7b8', 77679767, '84f026be-53e0-4759-ab89-131c4a66d649', true, false),
  ('65f65311-cc74-4c76-9cdf-29c5d674d40a', 70492884, '84f026be-53e0-4759-ab89-131c4a66d649', true, true),
  ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 39166759, '84f026be-53e0-4759-ab89-131c4a66d649', true, false),
  ('b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e', 55686346, '84f026be-53e0-4759-ab89-131c4a66d649', true, false),
  ('c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f', 81085720, '84f026be-53e0-4759-ab89-131c4a66d649', true, true),
  ('d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a', 52873146, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b', 75969632, '84f026be-53e0-4759-ab89-131c4a66d649', true, false);

-- Seed participants into the started event (Event #1)
INSERT INTO participants (id, membership_id, event_id) VALUES
  (1, '5d312426-ad56-4231-bb12-241acbfb91e2', 1),
  (2, 'c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d', 1),
  (3, 'b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e', 1);

-- Seed participants into the ended event (Event #2) with placements
INSERT INTO participants (id, membership_id, event_id, placement, signed_out_at) VALUES
  (4, 'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 2, 1, '2025-01-10 23:30:00'),
  (5, 'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e', 2, 2, '2025-01-10 23:25:00'),
  (6, 'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f', 2, 3, '2025-01-10 23:00:00'),
  (7, 'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a', 2, 4, '2025-01-10 22:30:00'),
  (8, 'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b', 2, 5, '2025-01-10 22:00:00');
SELECT setval('participants_num_seq', (SELECT MAX(id) FROM participants));

-- Seed rankings for members who participated in the ended event
-- Points calculated: ceil((payout * 5) / 50) * 1.0
INSERT INTO rankings (membership_id, points, attendance) VALUES
  ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 4, 1),
  ('b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e', 3, 1),
  ('c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f', 3, 1),
  ('d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a', 3, 1),
  ('e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b', 2, 1);