CREATE DATABASE IF NOT EXISTS adaptive_assessment CHARACTER SET utf8mb4;
USE adaptive_assessment;

CREATE TABLE users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(150) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role ENUM('CANDIDATE','ADMIN') NOT NULL DEFAULT 'CANDIDATE',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE topics (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE questions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  question_text TEXT NOT NULL,
  topic_id BIGINT UNSIGNED NOT NULL,
  category VARCHAR(50) NOT NULL,
  difficulty ENUM('EASY','MEDIUM','HARD') NOT NULL,
  question_type ENUM('MCQ','MULTIPLE_ANSWER','CODING','TRUE_FALSE') NOT NULL DEFAULT 'MCQ',
  explanation TEXT,
  time_limit_sec INT NOT NULL DEFAULT 60,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (topic_id) REFERENCES topics(id),
  INDEX idx_q_pick (active, difficulty, question_type),
  INDEX idx_q_topic (topic_id)
);

CREATE TABLE question_options (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  question_id BIGINT UNSIGNED NOT NULL,
  option_text VARCHAR(500) NOT NULL,
  is_correct BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE,
  INDEX idx_opt_q (question_id)
);

CREATE TABLE assessments (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(150) NOT NULL,
  description VARCHAR(500),
  assessment_type ENUM('MCQ','CODING') NOT NULL DEFAULT 'MCQ',
  total_questions INT NOT NULL DEFAULT 10,
  duration_minutes INT NOT NULL DEFAULT 30,
  show_explanations BOOLEAN NOT NULL DEFAULT FALSE,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE interview_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  candidate_id BIGINT UNSIGNED NOT NULL,
  assessment_id BIGINT UNSIGNED NOT NULL,
  status ENUM('IN_PROGRESS','COMPLETED','EXPIRED') NOT NULL DEFAULT 'IN_PROGRESS',
  current_difficulty ENUM('EASY','MEDIUM','HARD') NOT NULL DEFAULT 'MEDIUM',
  consecutive_correct INT NOT NULL DEFAULT 0,
  consecutive_wrong INT NOT NULL DEFAULT 0,
  answered_count INT NOT NULL DEFAULT 0,
  correct_count INT NOT NULL DEFAULT 0,
  total_score INT NOT NULL DEFAULT 0,
  started_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  completed_at DATETIME NULL,
  warnings INT NOT NULL DEFAULT 0,
  is_terminated BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (candidate_id) REFERENCES users(id),
  FOREIGN KEY (assessment_id) REFERENCES assessments(id),
  INDEX idx_sess_candidate (candidate_id, status)
);

CREATE TABLE session_questions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT UNSIGNED NOT NULL,
  question_id BIGINT UNSIGNED NOT NULL,
  order_no INT NOT NULL,
  difficulty ENUM('EASY','MEDIUM','HARD') NOT NULL,
  served_at DATETIME NOT NULL,
  answered_at DATETIME NULL,
  FOREIGN KEY (session_id) REFERENCES interview_sessions(id) ON DELETE CASCADE,
  FOREIGN KEY (question_id) REFERENCES questions(id),
  UNIQUE KEY uq_session_question (session_id, question_id)
);

CREATE TABLE responses (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_question_id BIGINT UNSIGNED NOT NULL UNIQUE,
  selected_option_id BIGINT UNSIGNED NOT NULL,
  is_correct BOOLEAN NOT NULL,
  points_awarded INT NOT NULL,
  time_taken_sec INT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_question_id) REFERENCES session_questions(id) ON DELETE CASCADE,
  FOREIGN KEY (selected_option_id) REFERENCES question_options(id)
);

CREATE TABLE interview_reports (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT UNSIGNED NOT NULL UNIQUE,
  candidate_id BIGINT UNSIGNED NOT NULL,
  assessment_id BIGINT UNSIGNED NOT NULL,
  total_questions INT NOT NULL,
  attempted INT NOT NULL,
  correct_count INT NOT NULL,
  accuracy DECIMAL(5,2) NOT NULL,
  score INT NOT NULL,
  max_score INT NOT NULL,
  avg_response_sec DECIMAL(8,2) NOT NULL,
  performance_level VARCHAR(30) NOT NULL,
  strong_topics TEXT NOT NULL,
  weak_topics TEXT NOT NULL,
  topic_breakdown TEXT NOT NULL,
  difficulty_breakdown TEXT NOT NULL,
  difficulty_progression TEXT NOT NULL,
  recommendations TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES interview_sessions(id) ON DELETE CASCADE,
  FOREIGN KEY (candidate_id) REFERENCES users(id),
  FOREIGN KEY (assessment_id) REFERENCES assessments(id),
  INDEX idx_report_candidate (candidate_id, created_at)
);

CREATE TABLE coding_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  candidate_id BIGINT UNSIGNED NOT NULL,
  assessment_id BIGINT UNSIGNED NOT NULL,
  status ENUM('IN_PROGRESS','COMPLETED','EXPIRED') NOT NULL DEFAULT 'IN_PROGRESS',
  score INT NOT NULL DEFAULT 0,
  started_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  completed_at DATETIME NULL,
  warnings INT NOT NULL DEFAULT 0,
  is_terminated BOOLEAN NOT NULL DEFAULT FALSE,
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