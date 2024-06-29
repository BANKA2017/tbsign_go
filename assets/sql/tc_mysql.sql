CREATE TABLE `tc_baiduid` (
  `id` int UNSIGNED NOT NULL,
  `uid` int UNSIGNED NOT NULL,
  `bduss` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `stoken` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `portrait` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_kd_growth` (
  `id` bigint NOT NULL,
  `uid` bigint NOT NULL,
  `pid` bigint NOT NULL,
  `status` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `log` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `date` int NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_options` (
  `name` varchar(124) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_plugins` (
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `ver` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `options` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_tieba` (
  `id` int UNSIGNED NOT NULL,
  `uid` int UNSIGNED NOT NULL,
  `pid` int UNSIGNED NOT NULL DEFAULT '0',
  `fid` int UNSIGNED NOT NULL DEFAULT '0',
  `tieba` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `no` tinyint(1) NOT NULL DEFAULT '0',
  `status` mediumint UNSIGNED NOT NULL DEFAULT '0',
  `latest` tinyint UNSIGNED NOT NULL DEFAULT '0',
  `last_error` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_users` (
  `id` int NOT NULL,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `pw` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `role` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'user',
  `t` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'tieba'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_users_options` (
  `uid` int NOT NULL,
  `name` varchar(124) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_ver4_ban_list` (
  `id` int NOT NULL,
  `uid` int NOT NULL,
  `pid` int NOT NULL,
  `name` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `name_show` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `portrait` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `tieba` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `stime` int NOT NULL,
  `etime` int NOT NULL,
  `log` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `date` int NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_ver4_ban_userset` (
  `uid` int NOT NULL,
  `c` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `tc_ver4_rank_log` (
  `id` int NOT NULL,
  `uid` int NOT NULL,
  `pid` int NOT NULL,
  `fid` int NOT NULL,
  `nid` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `tieba` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `log` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `date` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


ALTER TABLE `tc_baiduid`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uid_portrait` (`uid`,`portrait`),
  ADD UNIQUE KEY `id_uid` (`id`,`uid`) USING BTREE;

ALTER TABLE `tc_kd_growth`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id_uid_pid` (`id`,`uid`,`pid`),
  ADD KEY `uid` (`uid`),
  ADD KEY `pid` (`pid`),
  ADD KEY `date_id` (`date`,`id`) USING BTREE;

ALTER TABLE `tc_options`
  ADD UNIQUE KEY `name` (`name`);

ALTER TABLE `tc_plugins`
  ADD UNIQUE KEY `name` (`name`) USING BTREE;

ALTER TABLE `tc_tieba`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uid_pid_fid` (`uid`,`pid`,`fid`),
  ADD KEY `pid` (`pid`),
  ADD KEY `tieba_fid` (`tieba`,`fid`) USING BTREE;

ALTER TABLE `tc_users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`) USING BTREE,
  ADD UNIQUE KEY `email_name` (`email`,`name`) USING BTREE,
  ADD KEY `role` (`role`);

ALTER TABLE `tc_users_options`
  ADD UNIQUE KEY `uid_and_key_name` (`uid`,`name`) USING BTREE;

ALTER TABLE `tc_ver4_ban_list`
  ADD PRIMARY KEY (`id`),
  ADD KEY `uid` (`uid`),
  ADD KEY `id_uid` (`id`,`uid`),
  ADD KEY `pid` (`pid`),
  ADD KEY `id_date_stime_etime_uid` (`id`,`date`,`stime`,`etime`,`uid`) USING BTREE;

ALTER TABLE `tc_ver4_ban_userset`
  ADD UNIQUE KEY `uid` (`uid`);

ALTER TABLE `tc_ver4_rank_log`
  ADD PRIMARY KEY (`id`),
  ADD KEY `pid` (`pid`),
  ADD KEY `uid_pid` (`uid`,`pid`),
  ADD KEY `id_date` (`id`,`date`) USING BTREE;


ALTER TABLE `tc_baiduid`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

ALTER TABLE `tc_kd_growth`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT;

ALTER TABLE `tc_tieba`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

ALTER TABLE `tc_users`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

ALTER TABLE `tc_ver4_ban_list`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

ALTER TABLE `tc_ver4_rank_log`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;