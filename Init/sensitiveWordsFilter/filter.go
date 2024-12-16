package sensitiveWordsFilter

import (
	sensitive "github.com/pibuyu/sensitive_words_filter"
	"log"
)

func InitFilter() *sensitive.Manager {
	filter := sensitive.NewFilter()
	log.Println("----敏感词过滤器初始化完成----")
	return filter
}
