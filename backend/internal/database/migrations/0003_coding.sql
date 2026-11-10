ALTER TABLE assessments
  ADD COLUMN assessment_type ENUM('MCQ','CODING') NOT NULL DEFAULT 'MCQ' AFTER description;

CREATE TABLE coding_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  candidate_id BIGINT UNSIGNED NOT NULL,
  assessment_id BIGINT UNSIGNED NOT NULL,
  status ENUM('IN_PROGRESS','COMPLETED','EXPIRED') NOT NULL DEFAULT 'IN_PROGRESS',
  score INT NOT NULL DEFAULT 0,
  started_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  completed_at DATETIME NULL,
  FOREIGN KEY (candidate_id) REFERENCES users(id),
  FOREIGN KEY (assessment_id) REFERENCES assessments(id),
  INDEX idx_cs_candidate (candidate_id, status)
);

CREATE TABLE coding_session_problems (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT UNSIGNED NOT NULL,
  problem_slug VARCHAR(60) NOT NULL,
  order_no INT NOT NULL,
  status ENUM('OPEN','SOLVED','FAILED','SKIPPED') NOT NULL DEFAULT 'OPEN',
  attempts INT NOT NULL DEFAULT 0,
  failures INT NOT NULL DEFAULT 0,
  language VARCHAR(10) NOT NULL DEFAULT '',
  points INT NOT NULL DEFAULT 0,
  runtime_ms INT NOT NULL DEFAULT 0,
  solved_at DATETIME NULL,
  FOREIGN KEY (session_id) REFERENCES coding_sessions(id) ON DELETE CASCADE,
  UNIQUE KEY uq_csp_problem (session_id, problem_slug),
  UNIQUE KEY uq_csp_order (session_id, order_no)
);

CREATE TABLE coding_submissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_problem_id BIGINT UNSIGNED NOT NULL,
  language VARCHAR(10) NOT NULL,
  code MEDIUMTEXT NOT NULL,
  verdict VARCHAR(20) NOT NULL,
  passed INT NOT NULL DEFAULT 0,
  total INT NOT NULL DEFAULT 0,
  runtime_ms INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_problem_id) REFERENCES coding_session_problems(id) ON DELETE CASCADE,
  INDEX idx_sub_problem (session_problem_id, created_at)
);