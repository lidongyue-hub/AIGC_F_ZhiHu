package cron

import (
	"fmt"
	"qa/model"

	"github.com/robfig/cron"
)

func StartSchedule() {
	c := cron.New()

	// 每30分钟将redis数据同步到mysql
	addCronFunc(c, "@every 30m", func() {
		model.SyncUserLikeRecord()
		model.SyncAnswerLikeCount() // 有事务的运用，回滚
		model.FreeDeletedAnswersRecord()
	})

	// 每30分钟同步热榜信息
	addCronFunc(c, "@every 30m", func() {
		model.SyncHotQuestions() //这里面只是把近30天的所有问题ID及其标题还有其热度信息 都存在了缓存中，还未进行排序，
	})

	c.Start()
}

func addCronFunc(c *cron.Cron, sepc string, cmd func()) {
	err := c.AddFunc(sepc, cmd) // 是将一个特定的函数 cmd 根据给定的时间规则 spec 添加到 Cron 任务调度器中
	if err != nil {
		panic(fmt.Sprintf("定时任务异常: %v", err))
	}
}
