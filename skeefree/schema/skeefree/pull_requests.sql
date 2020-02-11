CREATE TABLE `pull_requests` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `org` varchar(128) NOT NULL,
  `repo` varchar(128) NOT NULL,
  `pull_request_number` bigint(20) unsigned NOT NULL,
  `title` varchar(4096) NOT NULL DEFAULT '',
  `author` varchar(128) NOT NULL DEFAULT '',
  `priority` int(1) NOT NULL DEFAULT '0',
  `status` varchar(32) NOT NULL,
  `is_open` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `submitted_by` varchar(128) NOT NULL,
  `added_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `probed_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `requested_review_by_db_reviewers` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `approved_by_db_reviewers` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `requested_review_by_db_infra` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `approved_by_db_infra` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `label_diff` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `label_detected` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `label_queued` tinyint(4) unsigned NOT NULL DEFAULT '0',
  `label_for_review` tinyint(4) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `org_repo_pull_idx` (`org`,`repo`,`pull_request_number`),
  KEY `status_idx` (`status`,`added_timestamp`),
  KEY `is_open_priority_idx` (`is_open`,`priority`)
) ENGINE=InnoDB AUTO_INCREMENT=154 DEFAULT CHARSET=utf8mb4;