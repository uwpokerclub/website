-- Seed the e2e testing user
INSERT INTO logins (username, password, role) VALUES ('e2e_user', '$2a$10$lzRaELvZxS2JwGsI0jSQueJWvMGfx82iYBuu0nFDCxuwJMabOHoX.', 'webmaster');

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

-- Seed an event for the seed semester
INSERT INTO events 
  (id, name, format, notes, semester_id, start_date, state, structure_id, rebuys, points_multiplier) 
VALUES
  (1, 'Winter 2025 Event #1', 'No Limit Hold''em', 'Seed event', '84f026be-53e0-4759-ab89-131c4a66d649', '2025-01-03 19:00:00', 0, 1, 0, 1.0);
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

-- Seed members into the seed semester
INSERT INTO memberships (id, user_id, semester_id, paid, discounted) VALUES
  ('5d312426-ad56-4231-bb12-241acbfb91e2', 62958169, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d', 20141158, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e', 85018940, '84f026be-53e0-4759-ab89-131c4a66d649', false, false),
  ('d4e5f6a7-b8c9-4d0e-a1b2-c3d4e5f6a7b8', 77679767, '84f026be-53e0-4759-ab89-131c4a66d649', true, false),
  ('65f65311-cc74-4c76-9cdf-29c5d674d40a', 70492884, '84f026be-53e0-4759-ab89-131c4a66d649', true, true);

-- Seed members into the seed event
INSERT INTO participants (id, membership_id, event_id) VALUES
  (1, '5d312426-ad56-4231-bb12-241acbfb91e2', 1),
  (2, 'c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d', 1),
  (3, 'b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e', 1);
SELECT setval('participants_id_seq', (SELECT MAX(id) FROM participants));