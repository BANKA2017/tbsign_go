BEGIN TRANSACTION;
-- DROP TABLE IF EXISTS "tc_ver4_rank_log";
-- CREATE TABLE IF NOT EXISTS "tc_ver4_rank_log" (
-- 	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
-- 	"uid"	INTEGER NOT NULL,
-- 	"pid"	INTEGER NOT NULL,
-- 	"fid"	INTEGER NOT NULL,
-- 	"nid"	TEXT NOT NULL,
-- 	"name"	TEXT NOT NULL,
-- 	"tieba"	TEXT NOT NULL,
-- 	"log"	TEXT COLLATE BINARY,
-- 	"date"	INTEGER NOT NULL
-- );
-- DROP TABLE IF EXISTS "tc_ver4_ban_userset";
-- CREATE TABLE IF NOT EXISTS "tc_ver4_ban_userset" (
-- 	"uid"	INTEGER NOT NULL UNIQUE,
-- 	"c"	TEXT COLLATE BINARY,
-- 	PRIMARY KEY("uid")
-- ) WITHOUT ROWID;
-- DROP TABLE IF EXISTS "tc_ver4_ban_list";
-- CREATE TABLE IF NOT EXISTS "tc_ver4_ban_list" (
-- 	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
-- 	"uid"	INTEGER NOT NULL,
-- 	"pid"	INTEGER NOT NULL,
-- 	"name"	TEXT COLLATE BINARY,
-- 	"name_show"	TEXT COLLATE BINARY,
-- 	"portrait"	TEXT COLLATE BINARY,
-- 	"tieba"	TEXT NOT NULL,
-- 	"stime"	INTEGER NOT NULL,
-- 	"etime"	INTEGER NOT NULL,
-- 	"log"	TEXT COLLATE BINARY,
-- 	"date"	INTEGER NOT NULL DEFAULT '0'
-- );
DROP TABLE IF EXISTS "tc_users_options";
CREATE TABLE IF NOT EXISTS "tc_users_options" (
	"uid"	INTEGER NOT NULL,
	"name"	TEXT NOT NULL,
	"value"	TEXT NOT NULL,
	UNIQUE("uid","name"),
	PRIMARY KEY("uid","name")
) WITHOUT ROWID;
DROP TABLE IF EXISTS "tc_users";
CREATE TABLE IF NOT EXISTS "tc_users" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	"name"	TEXT NOT NULL UNIQUE,
	"pw"	TEXT NOT NULL,
	"email"	TEXT NOT NULL,
	"role"	TEXT NOT NULL DEFAULT 'user',
	"t"	TEXT NOT NULL DEFAULT 'tieba',
	UNIQUE("email","name")
);
DROP TABLE IF EXISTS "tc_tieba";
CREATE TABLE IF NOT EXISTS "tc_tieba" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	"uid"	INTEGER NOT NULL,
	"pid"	INTEGER NOT NULL DEFAULT '0',
	"fid"	INTEGER NOT NULL DEFAULT '0',
	"tieba"	TEXT NOT NULL DEFAULT '',
	"no"	INTEGER NOT NULL DEFAULT '0',
	"status"	INTEGER NOT NULL DEFAULT '0',
	"latest"	INTEGER NOT NULL DEFAULT '0',
	"last_error"	TEXT COLLATE BINARY,
	UNIQUE("uid","pid","fid")
);
DROP TABLE IF EXISTS "tc_plugins";
CREATE TABLE IF NOT EXISTS "tc_plugins" (
	"name"	TEXT NOT NULL UNIQUE,
	"status"	INTEGER NOT NULL DEFAULT '0',
	"ver"	TEXT NOT NULL DEFAULT '',
	"options"	TEXT COLLATE BINARY,
	PRIMARY KEY("name")
) WITHOUT ROWID;
DROP TABLE IF EXISTS "tc_options";
CREATE TABLE IF NOT EXISTS "tc_options" (
	"name"	TEXT NOT NULL UNIQUE,
	"value"	TEXT NOT NULL,
	PRIMARY KEY("name")
) WITHOUT ROWID;
-- DROP TABLE IF EXISTS "tc_kd_growth";
-- CREATE TABLE IF NOT EXISTS "tc_kd_growth" (
-- 	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
-- 	"uid"	INTEGER NOT NULL,
-- 	"pid"	INTEGER NOT NULL,
-- 	"status"	TEXT COLLATE BINARY,
-- 	"log"	TEXT COLLATE BINARY,
-- 	"date"	INTEGER NOT NULL DEFAULT '0',
-- 	UNIQUE("id","uid","pid")
-- );
DROP TABLE IF EXISTS "tc_baiduid";
CREATE TABLE IF NOT EXISTS "tc_baiduid" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	"uid"	INTEGER NOT NULL,
	"bduss"	TEXT NOT NULL,
	"stoken"	TEXT NOT NULL,
	"name"	TEXT NOT NULL DEFAULT '',
	"portrait"	TEXT NOT NULL,
	UNIQUE("uid","portrait"),
	UNIQUE("id","uid")
);
DROP INDEX IF EXISTS "idx_tc_users_email_name";
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_users_email_name" ON "tc_users" (
	"email",
	"name"
);
DROP INDEX IF EXISTS "idx_tc_users_name";
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_users_name" ON "tc_users" (
	"name"
);
DROP INDEX IF EXISTS "idx_tc_tieba_uid_pid_fid";
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_tieba_uid_pid_fid" ON "tc_tieba" (
	"uid",
	"pid",
	"fid"
);
-- DROP INDEX IF EXISTS "idx_tc_kd_growth_id_uid_pid";
-- CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_kd_growth_id_uid_pid" ON "tc_kd_growth" (
-- 	"id",
-- 	"uid",
-- 	"pid"
-- );
DROP INDEX IF EXISTS "idx_tc_baiduid_uid_portraait";
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_baiduid_uid_portraait" ON "tc_baiduid" (
	"uid",
	"portrait"
);
DROP INDEX IF EXISTS "idx_tc_baiduid_id_uid";
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_baiduid_id_uid" ON "tc_baiduid" (
	"id",
	"uid"
);
-- DROP INDEX IF EXISTS "idx_tc_ver4_rank_log_id_date";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_rank_log_id_date" ON "tc_ver4_rank_log" (
-- 	"id",
-- 	"date"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_rank_log_uid_pid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_rank_log_uid_pid" ON "tc_ver4_rank_log" (
-- 	"uid",
-- 	"pid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_rank_log_pid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_rank_log_pid" ON "tc_ver4_rank_log" (
-- 	"pid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_ban_list_uid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_uid" ON "tc_ver4_ban_list" (
-- 	"uid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_ban_list_id_uid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_uid" ON "tc_ver4_ban_list" (
-- 	"id",
-- 	"uid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_ban_list_pid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_pid" ON "tc_ver4_ban_list" (
-- 	"pid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_ver4_ban_list_id_date_stime_etime_uid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_date_stime_etime_uid" ON "tc_ver4_ban_list" (
-- 	"id",
-- 	"date",
-- 	"stime",
-- 	"etime",
-- 	"uid"
-- );
DROP INDEX IF EXISTS "idx_tc_users_role";
CREATE INDEX IF NOT EXISTS "idx_tc_users_role" ON "tc_users" (
	"role"
);
DROP INDEX IF EXISTS "idx_tc_tieba_pid";
CREATE INDEX IF NOT EXISTS "idx_tc_tieba_pid" ON "tc_tieba" (
	"pid"
);
DROP INDEX IF EXISTS "idx_tc_tieba_tieba_fid";
CREATE INDEX IF NOT EXISTS "idx_tc_tieba_tieba_fid" ON "tc_tieba" (
	"tieba",
	"fid"
);
-- DROP INDEX IF EXISTS "idx_tc_kd_growth_date_id";
-- CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_date_id" ON "tc_kd_growth" (
-- 	"date",
-- 	"id"
-- );
-- DROP INDEX IF EXISTS "idx_tc_kd_growth_pid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_pid" ON "tc_kd_growth" (
-- 	"pid"
-- );
-- DROP INDEX IF EXISTS "idx_tc_kd_growth_uid";
-- CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_uid" ON "tc_kd_growth" (
-- 	"uid"
-- );
COMMIT;
