package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/qiniu/qshell/v2/iqshell/common/alert"
	"github.com/qiniu/qshell/v2/iqshell/common/data"
	"github.com/qiniu/qshell/v2/iqshell/common/log"
	"os"
	"strings"
)

type ListBucketInfo struct {
	CToken    string // aws continuation token
	Delimiter string
	MaxKeys   int64
	Prefix    string
	Id        string // id
	SecretKey string
	Region    string
	Bucket    string
}

func ListBucket(info ListBucketInfo) {

	if info.Id == "" || info.SecretKey == "" {
		log.Error(alert.CannotEmpty("AWS ID and SecretKey", ""))
		os.Exit(data.STATUS_ERROR)
	}

	if info.MaxKeys <= 0 || info.MaxKeys > 1000 {
		log.Warning("max key:%d out of range {}, change to 1000", info.MaxKeys)
		info.MaxKeys = 1000
	}

	// check AWS region
	if info.Region == "" {
		log.Error(alert.CannotEmpty("AWS region", ""))
		os.Exit(data.STATUS_ERROR)
	}

	// AWS related code
	s3session, err := session.NewSession()
	if err != nil {
		log.ErrorF("create AWS session error:%v", err)
		os.Exit(data.STATUS_ERROR)
	}
	s3session.Config.WithRegion(info.Region)
	s3session.Config.WithCredentials(credentials.NewStaticCredentials(info.Id, info.SecretKey, ""))

	svc := s3.New(s3session)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(info.Bucket),
		Prefix:  aws.String(info.Prefix),
		MaxKeys: aws.Int64(info.MaxKeys),
	}

	if info.CToken != "" {
		input.ContinuationToken = aws.String(info.CToken)
	}

	for {
		result, err := svc.ListObjectsV2(input)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					log.ErrorF("list error:%s error:%v", s3.ErrCodeNoSuchBucket, aerr.Error())
				default:
					log.ErrorF("list error:%v", aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.ErrorF("list error:%v", aerr.Error())
			}
			log.ErrorF("ContinuationToken: %v", input.ContinuationToken)
			os.Exit(data.STATUS_ERROR)
		}

		for _, obj := range result.Contents {
			if strings.HasSuffix(*obj.Key, "/") && *obj.Size == 0 { // 跳过目录
				continue
			}
			log.AlertF("%s\t%d\t%s\t%s\n", *obj.Key, *obj.Size, *obj.ETag, *obj.LastModified)
		}

		if *result.IsTruncated {
			input.ContinuationToken = result.NextContinuationToken
		} else {
			break
		}
	}
}
