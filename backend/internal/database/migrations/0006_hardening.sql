-- Token versioning (JWT revocation) and per-visit demo accounts
ALTER TABLE users
  ADD COLUMN token_version INT NOT NULL DEFAULT 0,
  ADD COLUMN is_demo BOOLEAN NOT NULL DEFAULT FALSE,
  ADD INDEX idx_users_demo (is_demo, created_at);

-- Close duplicate live attempts left over from before this rule (the newest one survives)
UPDATE interview_sessions s
JOIN (
  SELECT candidate_id, assessment_id, MAX(id) AS keep_id
  FROM interview_sessions WHERE status = 'IN_PROGRESS'
  GROUP BY candidate_id, assessment_id HAVING COUNT(*) > 1
) d ON d.candidate_id = s.candidate_id AND d.assessment_id = s.assessment_id
SET s.status = 'EXPIRED', s.completed_at = NOW()
WHERE s.status = 'IN_PROGRESS' AND s.id <> d.keep_id;

UPDATE coding_sessions s
JOIN (
  SELECT candidate_id, assessment_id, MAX(id) AS keep_id
  FROM coding_sessions WHERE status = 'IN_PROGRESS'
  GROUP BY candidate_id, assessment_id HAVING COUNT(*) > 1
) d ON d.candidate_id = s.candidate_id AND d.assessment_id = s.assessment_id
SET s.status = 'EXPIRED', s.completed_at = NOW()
WHERE s.status = 'IN_PROGRESS' AND s.id <> d.keep_id;

-- MySQL has no partial unique index, so a generated column is NULL unless the attempt is live
ALTER TABLE interview_sessions
  ADD COLUMN active_assessment_id BIGINT UNSIGNED
    GENERATED ALWAYS AS (IF(status = 'IN_PROGRESS', assessment_id, NULL)) STORED,
  ADD UNIQUE KEY uq_one_live_session (candidate_id, active_assessment_id);

ALTER TABLE coding_sessions
  ADD COLUMN active_assessment_id BIGINT UNSIGNED
    GENERATED ALWAYS AS (IF(status = 'IN_PROGRESS', assessment_id, NULL)) STORED,
  ADD UNIQUE KEY uq_one_live_coding (candidate_id, active_assessment_id);