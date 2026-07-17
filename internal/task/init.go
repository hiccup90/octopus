package task

import (
	"context"
	"time"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
	"github.com/bestruirui/octopus/internal/utils/log"
)

const (
	TaskStatsSave    = "stats_save"
	TaskRelayLogSave = "relay_log_save"
	TaskSyncLLM      = "sync_llm"
	TaskCleanLLM     = "clean_llm"
	TaskBaseUrlDelay = "base_url_delay"
)

func Init() {
	// 注册基础URL延迟任务
	Register(TaskBaseUrlDelay, 1*time.Hour, true, ChannelBaseUrlDelayTask)

	// 注册LLM同步任务（失败不影响后续任务注册）
	if syncLLMIntervalHours, err := op.SettingGetInt(model.SettingKeySyncLLMInterval); err != nil {
		log.Warnf("failed to get sync LLM interval: %v", err)
	} else {
		syncLLMInterval := time.Duration(syncLLMIntervalHours) * time.Hour
		Register(string(model.SettingKeySyncLLMInterval), syncLLMInterval, true, SyncModelsTask)
	}

	// 注册统计保存任务（失败不影响后续任务注册）
	if statsSaveIntervalMinutes, err := op.SettingGetInt(model.SettingKeyStatsSaveInterval); err != nil {
		log.Warnf("failed to get stats save interval: %v", err)
	} else {
		statsSaveInterval := time.Duration(statsSaveIntervalMinutes) * time.Minute
		Register(TaskStatsSave, statsSaveInterval, false, op.StatsSaveDBTask)
	}

	// 注册中继日志保存任务
	Register(TaskRelayLogSave, 10*time.Minute, false, func() {
		if err := op.RelayLogSaveDBTask(context.Background()); err != nil {
			log.Warnf("relay log save db task failed: %v", err)
		}
	})
}
