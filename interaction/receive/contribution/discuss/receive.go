package discuss

import (
	"haifengonline/models/common"
)

type GetDiscussVideoListReceiveStruct struct {
	PageInfo common.PageInfo `json:"page_info"`
}

type GetDiscussArticleListReceiveStruct struct {
	PageInfo common.PageInfo `json:"page_info"`
}

type GetDiscussBarrageListReceiveStruct struct {
	PageInfo common.PageInfo `json:"page_info"`
}
