package entity

// ScriptScore 脚本评分
type ScriptScore struct {
	ID       int64 `gorm:"column:id" json:"id" form:"id"`
	UserId   int64 `gorm:"column:user_id;index:user_script,unique;index:user" json:"user_id" form:"user_id"`
	ScriptId int64 `gorm:"column:script_id;index:user_script,unique;index:script" json:"script_id" form:"script_id"`
	// 评分,五星制,50
	Score int64 `gorm:"column:score" json:"score" form:"score"`
	// 评分原因
	Message    string `gorm:"column:message;type:text" json:"message" form:"message"`
	State      int32  `gorm:"column:state;type:int(10);default:1" json:"state" form:"state"`
	Createtime int64  `gorm:"column:createtime" json:"createtime" form:"createtime"`
	Updatetime int64  `gorm:"column:updatetime" json:"updatetime" form:"updatetime"`
}

// ScriptStatistics 脚本总下载更新统计
type ScriptStatistics struct {
	ID         int64 `gorm:"column:id" json:"id" form:"id"`
	ScriptId   int64 `gorm:"column:script_id;index:script,unique" json:"script_id" form:"script_id"`
	Download   int64 `gorm:"column:download;default:0" json:"download" form:"download"`
	Update     int64 `gorm:"column:update;default:0" json:"update" form:"update"`
	Score      int64 `gorm:"column:score;default:0" json:"score" form:"score"`
	ScoreCount int64 `gorm:"column:score_count;default:0" json:"score_count" form:"score_count"`
}

// ScriptDateStatistics 脚本日下载更新统计
type ScriptDateStatistics struct {
	ID       int64  `gorm:"column:id" json:"id" form:"id"`
	ScriptId int64  `gorm:"column:script_id;index:script_date,unique;default:0" json:"script_id" form:"script_id"`
	Date     string `gorm:"type:varchar(255);column:date;index:script_date,unique;default:0" json:"date" form:"date"`
	Download int64  `gorm:"column:download;default:0" json:"download" form:"download"`
	Update   int64  `gorm:"column:update;default:0" json:"update" form:"update"`
}
