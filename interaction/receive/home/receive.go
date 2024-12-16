package home

import (
	"haifengonline/models/common"
)

type GetHomeInfoReceiveStruct struct {
	PageInfo common.PageInfo `json:"page_info" binding:"required"`
}

type SubmitBugReceiveStruct struct {
	Content string `json:"content" binding:"required"`
	Phone   string `json:"phone"`
}
