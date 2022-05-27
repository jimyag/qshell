package operations

import (
	"github.com/qiniu/qshell/v2/iqshell"
	"github.com/qiniu/qshell/v2/iqshell/common/alert"
	"github.com/qiniu/qshell/v2/iqshell/common/data"
	"github.com/qiniu/qshell/v2/iqshell/common/flow"
	"github.com/qiniu/qshell/v2/iqshell/common/log"
	"github.com/qiniu/qshell/v2/iqshell/storage/object/batch"
	"github.com/qiniu/qshell/v2/iqshell/storage/object/download"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultDeadline = 3600
)

type PrivateUrlInfo struct {
	PublicUrl string
	Deadline  string
}

func (p PrivateUrlInfo) WorkId() string {
	return p.PublicUrl
}

func (p *PrivateUrlInfo) Check() *data.CodeError {
	if len(p.PublicUrl) == 0 {
		return alert.CannotEmptyError("PublicUrl", "")
	}
	return nil
}

func (p PrivateUrlInfo) getDeadlineOfInt() (int64, *data.CodeError) {
	if len(p.Deadline) == 0 {
		return time.Now().Add(time.Second * DefaultDeadline).Unix(), nil
	}

	if val, err := strconv.ParseInt(p.Deadline, 10, 64); err != nil {
		return 0, data.NewEmptyError().AppendDesc("invalid deadline")
	} else {
		return val, nil
	}
}

func PrivateUrl(cfg *iqshell.Config, info PrivateUrlInfo) {
	if shouldContinue := iqshell.CheckAndLoad(cfg, iqshell.CheckAndLoadInfo{
		Checker: &info,
	}); !shouldContinue {
		return
	}

	deadline, err := info.getDeadlineOfInt()
	if err != nil {
		log.Error(err)
		return
	}

	url, err := download.PublicUrlToPrivate(download.PublicUrlToPrivateApiInfo{
		PublicUrl: info.PublicUrl,
		Deadline:  deadline,
	})

	log.Alert(url)
}

type BatchPrivateUrlInfo struct {
	BatchInfo batch.Info
	Deadline  string
}

func (info *BatchPrivateUrlInfo) Check() *data.CodeError {
	return nil
}

// BatchPrivateUrl 批量删除，由于和批量删除的输入读取逻辑不同，所以分开
func BatchPrivateUrl(cfg *iqshell.Config, info BatchPrivateUrlInfo) {
	if shouldContinue := iqshell.CheckAndLoad(cfg, iqshell.CheckAndLoadInfo{
		Checker: &info,
	}); !shouldContinue {
		return
	}

	flow.New(info.BatchInfo.Info).
		WorkProviderWithFile(info.BatchInfo.InputFile,
			info.BatchInfo.EnableStdin,
			flow.NewItemsWorkCreator(info.BatchInfo.ItemSeparate, 1, func(items []string) (work flow.Work, err *data.CodeError) {
				url := items[0]
				if url == "" {
					return nil, alert.Error("url invalid", "")
				}

				urlToSign := strings.TrimSpace(url)
				if urlToSign == "" {
					return nil, alert.Error("url invalid after TrimSpace", "")
				}
				return &PrivateUrlInfo{
					PublicUrl: url,
					Deadline:  info.Deadline,
				}, nil
			})).
		WorkerProvider(flow.NewWorkerProvider(func() (flow.Worker, *data.CodeError) {
			return flow.NewSimpleWorker(func(workInfo *flow.WorkInfo) (flow.Result, *data.CodeError) {
				in := workInfo.Work.(*PrivateUrlInfo)
				if deadline, gErr := in.getDeadlineOfInt(); gErr == nil {
					if r, pErr := download.PublicUrlToPrivate(download.PublicUrlToPrivateApiInfo{
						PublicUrl: in.PublicUrl,
						Deadline:  deadline,
					}); pErr != nil {
						return nil, pErr
					} else {
						return r, nil
					}
				} else {
					return nil, gErr
				}
			}), nil
		})).
		OnWorkSuccess(func(work *flow.WorkInfo, result flow.Result) {
			url, _ := result.(string)
			log.Alert(url)
		}).
		OnWorkFail(func(work *flow.WorkInfo, err *data.CodeError) {
			log.Error(err)
		}).Build().Start()
}