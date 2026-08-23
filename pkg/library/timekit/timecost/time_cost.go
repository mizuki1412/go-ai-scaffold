package timecost

import (
	"fmt"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"time"
)

// TimeCost 程序执行耗时记录打点
type TimeCost struct {
	Title string
	Start time.Time
}

type Params struct{}

func NewTimeCost(title string, params ...Params) TimeCost {
	return TimeCost{
		Title: title,
		Start: time.Now(),
	}
}

// 每次打印后重新计时
func (th *TimeCost) PrintCost(msg string, logInfo ...bool) {
	message := fmt.Sprintf("time-cost【%s|%s】%fs", th.Title, msg, time.Since(th.Start).Seconds())
	if len(logInfo) > 0 && logInfo[0] {
		logkit.Info(message)
	} else {
		logkit.Debug(message)
	}
	th.Start = time.Now()
}
