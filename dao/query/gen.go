// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:               db,
		TcBaiduid:        newTcBaiduid(db, opts...),
		TcKdGrowth:       newTcKdGrowth(db, opts...),
		TcOption:         newTcOption(db, opts...),
		TcPlugin:         newTcPlugin(db, opts...),
		TcTieba:          newTcTieba(db, opts...),
		TcUser:           newTcUser(db, opts...),
		TcUsersOption:    newTcUsersOption(db, opts...),
		TcVer4BanList:    newTcVer4BanList(db, opts...),
		TcVer4BanUserset: newTcVer4BanUserset(db, opts...),
		TcVer4RankLog:    newTcVer4RankLog(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	TcBaiduid        tcBaiduid
	TcKdGrowth       tcKdGrowth
	TcOption         tcOption
	TcPlugin         tcPlugin
	TcTieba          tcTieba
	TcUser           tcUser
	TcUsersOption    tcUsersOption
	TcVer4BanList    tcVer4BanList
	TcVer4BanUserset tcVer4BanUserset
	TcVer4RankLog    tcVer4RankLog
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:               db,
		TcBaiduid:        q.TcBaiduid.clone(db),
		TcKdGrowth:       q.TcKdGrowth.clone(db),
		TcOption:         q.TcOption.clone(db),
		TcPlugin:         q.TcPlugin.clone(db),
		TcTieba:          q.TcTieba.clone(db),
		TcUser:           q.TcUser.clone(db),
		TcUsersOption:    q.TcUsersOption.clone(db),
		TcVer4BanList:    q.TcVer4BanList.clone(db),
		TcVer4BanUserset: q.TcVer4BanUserset.clone(db),
		TcVer4RankLog:    q.TcVer4RankLog.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:               db,
		TcBaiduid:        q.TcBaiduid.replaceDB(db),
		TcKdGrowth:       q.TcKdGrowth.replaceDB(db),
		TcOption:         q.TcOption.replaceDB(db),
		TcPlugin:         q.TcPlugin.replaceDB(db),
		TcTieba:          q.TcTieba.replaceDB(db),
		TcUser:           q.TcUser.replaceDB(db),
		TcUsersOption:    q.TcUsersOption.replaceDB(db),
		TcVer4BanList:    q.TcVer4BanList.replaceDB(db),
		TcVer4BanUserset: q.TcVer4BanUserset.replaceDB(db),
		TcVer4RankLog:    q.TcVer4RankLog.replaceDB(db),
	}
}

type queryCtx struct {
	TcBaiduid        *tcBaiduidDo
	TcKdGrowth       *tcKdGrowthDo
	TcOption         *tcOptionDo
	TcPlugin         *tcPluginDo
	TcTieba          *tcTiebaDo
	TcUser           *tcUserDo
	TcUsersOption    *tcUsersOptionDo
	TcVer4BanList    *tcVer4BanListDo
	TcVer4BanUserset *tcVer4BanUsersetDo
	TcVer4RankLog    *tcVer4RankLogDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		TcBaiduid:        q.TcBaiduid.WithContext(ctx),
		TcKdGrowth:       q.TcKdGrowth.WithContext(ctx),
		TcOption:         q.TcOption.WithContext(ctx),
		TcPlugin:         q.TcPlugin.WithContext(ctx),
		TcTieba:          q.TcTieba.WithContext(ctx),
		TcUser:           q.TcUser.WithContext(ctx),
		TcUsersOption:    q.TcUsersOption.WithContext(ctx),
		TcVer4BanList:    q.TcVer4BanList.WithContext(ctx),
		TcVer4BanUserset: q.TcVer4BanUserset.WithContext(ctx),
		TcVer4RankLog:    q.TcVer4RankLog.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
