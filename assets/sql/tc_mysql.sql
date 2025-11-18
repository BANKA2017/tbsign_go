CREATE TABLE IF NOT EXISTS `tc_baiduid` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` int UNSIGNED NOT NULL,
  `bduss` mediumtext COLLATE utf8mb4_general_ci NOT NULL,
  `stoken` text COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL,
  `portrait` varchar(40) COLLATE utf8mb4_general_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_uid` (`id`,`uid`),
  UNIQUE KEY `uid_portrait` (`uid`,`portrait`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `tc_options` (
  `name` varchar(124) COLLATE utf8mb4_general_ci NOT NULL,
  `value` text COLLATE utf8mb4_general_ci NOT NULL,
  PRIMARY KEY (`name`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `tc_plugins` (
  `name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `status` tinyint(1) NOT NULL,
  `ver` varchar(15) COLLATE utf8mb4_general_ci NOT NULL,
  `options` mediumtext COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`name`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `tc_tieba` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` int UNSIGNED NOT NULL,
  `pid` int UNSIGNED NOT NULL,
  `fid` int UNSIGNED NOT NULL,
  `tieba` varchar(200) COLLATE utf8mb4_general_ci NOT NULL,
  `no` tinyint(1) NOT NULL,
  `status` mediumint UNSIGNED NOT NULL,
  `latest` tinyint UNSIGNED NOT NULL,
  `last_error` text COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_pid_fid` (`uid`,`pid`,`fid`),
  KEY `pid` (`pid`),
  KEY `tieba_fid` (`tieba`,`fid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `tc_users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(20) COLLATE utf8mb4_general_ci NOT NULL,
  `pw` text COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(40) COLLATE utf8mb4_general_ci NOT NULL,
  `role` varchar(10) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'user',
  `t` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'tieba',
  PRIMARY KEY (`id`),
  UNIQUE KEY `email_name` (`email`,`name`),
  UNIQUE KEY `name` (`name`),
  KEY `role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `tc_users_options` (
  `uid` int NOT NULL,
  `name` varchar(124) COLLATE utf8mb4_general_ci NOT NULL,
  `value` text COLLATE utf8mb4_general_ci NOT NULL,
  PRIMARY KEY (`uid`,`name`),
  UNIQUE KEY `uid_and_key_name` (`uid`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
