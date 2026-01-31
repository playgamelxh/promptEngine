CREATE TABLE IF NOT EXISTS `llm_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `api_key` longtext,
  `base_url` longtext,
  `model_name` longtext,
  `temperature` double DEFAULT '0.7',
  `tags` longtext,
  `is_default` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_llm_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `projects` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `description` longtext,
  `tags` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_projects_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `prompts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `name` longtext,
  `content` longtext,
  `tags` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_prompts_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `test_cases` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `prompt_id` bigint unsigned DEFAULT NULL,
  `input` longtext,
  `input_md5` varchar(32) DEFAULT NULL,
  `expected_output` longtext,
  `tags` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_test_cases_deleted_at` (`deleted_at`),
  KEY `idx_test_cases_project_id` (`project_id`),
  KEY `idx_test_cases_input_md5` (`input_md5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `llm_test_cases` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `prompt_id` bigint unsigned DEFAULT NULL,
  `input` longtext,
  `output` longtext,
  `evaluation` longtext,
  `is_pass` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_llm_test_cases_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
