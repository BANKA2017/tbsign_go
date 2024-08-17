BEGIN TRANSACTION;
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
COMMIT;
