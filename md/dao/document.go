package dao

import (
	"md/model/common"
	"md/model/entity"
	"md/util"
	"sort"

	"github.com/jmoiron/sqlx"
)

// 添加文档
func DocumentAdd(tx *sqlx.Tx, document entity.Document) error {
	sql := `insert into t_document (id,name,content,type,published,create_time,update_time,book_id,user_id) values (:id,:name,:content,:type,:published,:create_time,:update_time,:book_id,:user_id)`
	_, err := tx.NamedExec(sql, document)
	return err
}

// 修改文档基础信息
func DocumentUpdate(tx *sqlx.Tx, document entity.Document) error {
	sql := `update t_document set name=:name,published=:published,book_id=:book_id where id=:id and user_id=:user_id`
	_, err := tx.NamedExec(sql, document)
	return err
}

// 修改文档内容
func DocumentUpdateContent(tx *sqlx.Tx, document entity.Document) error {
	sql := `update t_document set content=:content,update_time=:update_time where id=:id and user_id=:user_id`
	_, err := tx.NamedExec(sql, document)
	return err
}

// 根据id删除文档
func DocumentDeleteById(tx *sqlx.Tx, id, userId string) error {
	sql := `delete from t_document where id=$1 and user_id=$2`
	_, err := tx.Exec(sql, id, userId)
	return err
}

// 查询文档列表
func DocumentList(db *sqlx.DB, bookId, userId string) ([]entity.Document, error) {
	sqlCompletion := util.SqlCompletion{}
	sqlCompletion.InitSql(`select id,name,type,published,create_time,update_time,book_id from t_document`)
	sqlCompletion.Eq("user_id", userId, true)
	if bookId != "" {
		sqlCompletion.Eq("book_id", bookId, true)
	}
	result := []entity.Document{}
	err := db.Select(&result, sqlCompletion.GetSql(), sqlCompletion.GetParams()...)
	// 按名称升序
	sort.Slice(result, func(i, j int) bool {
		return util.StringSort(result[i].Name, result[j].Name)
	})
	return result, err
}

// 根据id查询文档
func DocumentGetById(db *sqlx.DB, id, userId string) (entity.Document, error) {
	sql := `select id,name,content,type,published,create_time,update_time,book_id from t_document where id=$1 and user_id=$2`
	result := entity.Document{}
	err := db.Get(&result, sql, id, userId)
	return result, err
}

// 清空文档的bookId
func DocumentClearBookId(tx *sqlx.Tx, bookId string) error {
	sql := `update t_document set book_id='' where book_id=$1`
	_, err := tx.Exec(sql, bookId)
	return err
}

// 根据id查询公开发布文档
func DocumentGetPublished(db *sqlx.DB, id string) (entity.Document, error) {
	sql := `select id,name,content,type,published,create_time,update_time,book_id from t_document where id=$1 and published=true`
	result := entity.Document{}
	err := db.Get(&result, sql, id)
	return result, err
}

// 分页查询公开发布文档列表
func DocumentPagePublished(db *sqlx.DB, pageCondition common.PageCondition[entity.DocumentPageCondition]) ([]entity.DocumentPageResult, int, error) {
	sqlCompletion := util.SqlCompletion{}
	sqlCompletion.InitSql(
		`select a.id, a.name, a.type, a.create_time, a.update_time, COALESCE(b.name, '') as username, COALESCE(c.name, '') as book_name 
		from t_document a 
		left join t_user b on a.user_id = b.id 
		left join t_book c on a.book_id = c.id`,
	)
	sqlCompletion.Eq("a.published", true, true)
	if pageCondition.Condition.Username != "" {
		sqlCompletion.Like("b.name", pageCondition.Condition.Username, true)
	}
	if pageCondition.Condition.Name != "" {
		sqlCompletion.Like("a.name", pageCondition.Condition.Name, true)
	}
	if pageCondition.Condition.Type != "" {
		sqlCompletion.Like("a.type", pageCondition.Condition.Type, true)
	}
	if pageCondition.Condition.BookName != "" {
		sqlCompletion.Like("c.name", pageCondition.Condition.BookName, true)
	}
	sqlCompletion.Order("a.create_time", false)
	sqlCompletion.Limit(pageCondition.Page.Current, pageCondition.Page.Size)

	// 查询分页数据
	result := []entity.DocumentPageResult{}
	err := db.Select(&result, sqlCompletion.GetSql(), sqlCompletion.GetParams()...)
	if err != nil {
		return result, 0, err
	}

	// 查询总记录数
	countResult := common.CountResult{}
	err = db.Get(&countResult, sqlCompletion.GetCountSql(), sqlCompletion.GetCountParams()...)
	if err != nil {
		return result, 0, err
	}

	return result, countResult.Count, err
}
