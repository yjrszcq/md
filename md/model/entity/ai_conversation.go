package entity

// AIConversation AI对话实体
type AIConversation struct {
	Id         string `json:"id" db:"id"`
	UserId     string `json:"userId" db:"user_id"`
	Title      string `json:"title" db:"title"`
	Content    string `json:"content" db:"content"` // JSON存储对话内容
	CreateTime int64  `json:"createTime" db:"create_time"`
	UpdateTime int64  `json:"updateTime" db:"update_time"`
}

// AIConversationListItem 对话列表项（不含完整内容）
type AIConversationListItem struct {
	Id         string `json:"id" db:"id"`
	Title      string `json:"title" db:"title"`
	CreateTime int64  `json:"createTime" db:"create_time"`
	UpdateTime int64  `json:"updateTime" db:"update_time"`
}

// AIConversationCondition 对话请求条件
type AIConversationCondition struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Keyword string `json:"keyword"` // 搜索关键词
}
