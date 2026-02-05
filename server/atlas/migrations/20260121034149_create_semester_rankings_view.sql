-- Migration: Create semester_rankings_view with RANK() window function
--
-- Purpose: Compute member rankings with proper tie handling using RANK()
-- instead of ROW_NUMBER(). RANK() assigns the same position to tied scores
-- and skips subsequent positions (e.g., 1, 1, 3 instead of 1, 2, 3).
--
-- Ordering: RANK() is computed based on points only. Alphabetical ordering
-- for display purposes should be applied at the query level (ORDER BY position,
-- last_name, first_name).
--
-- Performance: Standard view (non-materialized), RANK() computed on every query.
-- Requires index on memberships.semester_id for optimal performance.
--
-- Usage: Query via service methods (GetRanking, GetRankings) - do not query directly.

CREATE VIEW semester_rankings_view AS
SELECT
    m.semester_id,
    m.id AS membership_id,
    u.id AS user_id,
    u.first_name,
    u.last_name,
    r.points,
    RANK() OVER (
        PARTITION BY m.semester_id
        ORDER BY r.points DESC NULLS LAST
    ) AS position
FROM memberships m
INNER JOIN rankings r ON m.id = r.membership_id
INNER JOIN users u ON m.user_id = u.id;
