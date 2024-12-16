package date

import (
	"haifengonline/consts"
	"strconv"
	"time"
)

func GetDay(time time.Time) int {
	ret, _ := strconv.Atoi(time.Format(consts.TIME_DAY))
	return ret
}

func GetYesterday() int {
	ret, _ := strconv.Atoi(time.Now().Add(-1 * 24 * time.Hour).Format(consts.TIME_DAY))
	return ret
}
