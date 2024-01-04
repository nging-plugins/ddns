package dnsdomain

// UpdateStatusType 更新状态
type UpdateStatusType string

const (
	// UpdatedNothing 未改变
	UpdatedNothing UpdateStatusType = "未改变"
	// UpdatedFailed 更新失败
	UpdatedFailed UpdateStatusType = "失败"
	// UpdatedSuccess 更新成功
	UpdatedSuccess UpdateStatusType = "成功"
	// UpdatedIdle 待更新
	UpdatedIdle UpdateStatusType = "待更新"
)
