package operations

import (
	"github.com/qiniu/qshell/v2/iqshell"
	"github.com/qiniu/qshell/v2/iqshell/common/alert"
	"github.com/qiniu/qshell/v2/iqshell/common/data"
	"github.com/qiniu/qshell/v2/iqshell/common/export"
	"github.com/qiniu/qshell/v2/iqshell/common/flow"
	"github.com/qiniu/qshell/v2/iqshell/common/log"
	"github.com/qiniu/qshell/v2/iqshell/storage/object"
	"github.com/qiniu/qshell/v2/iqshell/storage/object/batch"
)

type MoveInfo object.MoveApiInfo

func (info *MoveInfo) Check() *data.CodeError {
	if len(info.SourceBucket) == 0 {
		return alert.CannotEmptyError("SourceBucket", "")
	}
	if len(info.SourceKey) == 0 {
		return alert.CannotEmptyError("SourceKey", "")
	}
	if len(info.DestBucket) == 0 {
		return alert.CannotEmptyError("DestBucket", "")
	}
	if len(info.DestKey) == 0 {
		return alert.CannotEmptyError("DestKey", "")
	}
	return nil
}

func Move(cfg *iqshell.Config, info MoveInfo) {
	if shouldContinue := iqshell.CheckAndLoad(cfg, iqshell.CheckAndLoadInfo{
		Checker: &info,
	}); !shouldContinue {
		return
	}

	result, err := object.Move((*object.MoveApiInfo)(&info))
	if err != nil {
		log.ErrorF("Move Failed, [%s:%s] => [%s:%s], Error: %v",
			info.SourceBucket, info.SourceKey,
			info.DestBucket, info.DestKey,
			err)
		return
	}

	if len(result.Error) != 0 {
		log.ErrorF("Move Failed, [%s:%s] => [%s:%s], Code: %d, Error: %s",
			info.SourceBucket, info.SourceKey,
			info.DestBucket, info.DestKey,
			result.Code, result.Error)
		return
	}

	if result.IsSuccess() {
		log.InfoF("Move Success, [%s:%s] => [%s:%s]",
			info.SourceBucket, info.SourceKey,
			info.DestBucket, info.DestKey)
	}
}

type BatchMoveInfo struct {
	BatchInfo    batch.Info
	SourceBucket string
	DestBucket   string
}

func (info *BatchMoveInfo) Check() *data.CodeError {
	if err := info.BatchInfo.Check(); err != nil {
		return err
	}

	if len(info.SourceBucket) == 0 {
		return alert.CannotEmptyError("SrcBucket", "")
	}

	if len(info.DestBucket) == 0 {
		return alert.CannotEmptyError("DestBucket", "")
	}

	return nil
}

func BatchMove(cfg *iqshell.Config, info BatchMoveInfo) {
	if shouldContinue := iqshell.CheckAndLoad(cfg, iqshell.CheckAndLoadInfo{
		Checker: &info,
	}); !shouldContinue {
		return
	}

	exporter, err := export.NewFileExport(info.BatchInfo.FileExporterConfig)
	if err != nil {
		log.Error(err)
		return
	}

	batch.NewHandler(info.BatchInfo).ItemsToOperation(func(items []string) (operation batch.Operation, err *data.CodeError) {
		srcKey, destKey := items[0], items[0]
		if len(items) > 1 {
			destKey = items[1]
		}
		if srcKey != "" && destKey != "" {
			return &object.MoveApiInfo{
				SourceBucket: info.SourceBucket,
				SourceKey:    srcKey,
				DestBucket:   info.DestBucket,
				DestKey:      destKey,
				Force:        info.BatchInfo.Force,
			}, nil
		}
		return nil, alert.Error("key invalid", "")
	}).OnResult(func(operationInfo string, operation batch.Operation, result *batch.OperationResult) {
		apiInfo, ok := (operation).(*object.MoveApiInfo)
		if !ok {
			return
		}

		if result.Code != 200 || result.Error != "" {
			exporter.Fail().ExportF("%s%s%d-%s", operationInfo, flow.ErrorSeparate, result.Code, result.Error)
			log.ErrorF("Move Failed, [%s:%s] => [%s:%s], Code: %d, Error: %s",
				apiInfo.SourceBucket, apiInfo.SourceKey,
				apiInfo.DestBucket, apiInfo.DestKey,
				result.Code, result.Error)
		} else {
			exporter.Success().Export(operationInfo)
			log.InfoF("Move Success, [%s:%s] => [%s:%s]",
				apiInfo.SourceBucket, apiInfo.SourceKey,
				apiInfo.DestBucket, apiInfo.DestKey)
		}
	}).OnError(func(err *data.CodeError) {
		log.ErrorF("Batch move error:%v:", err)
	}).Start()
}
