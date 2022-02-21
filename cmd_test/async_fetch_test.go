package cmd

import (
	"github.com/qiniu/qshell/v2/cmd_test/test"
	"strings"
	"testing"
)

func TestAsyncFetch(t *testing.T) {
	content := test.BucketObjectDomainsString
	content += "https://qshell-na0.qiniupkg.com/hello10.json"
	path, err := test.CreateFileWithContent("async_fetch.txt", content)
	if err != nil {
		t.Fatal("create path error:", err)
	}

	successLogPath, err := test.CreateFileWithContent("async_fetch_success_log.txt", "")
	if err != nil {
		t.Fatal("create successLogPath error:", err)
	}

	failLogPath, err := test.CreateFileWithContent("async_fetch_fail_log.txt", "")
	if err != nil {
		t.Fatal("create failLogPath error:", err)
	}

	test.RunCmdWithError("abfetch", test.Bucket,
		"-i", path,
		"-s", successLogPath,
		"-e", failLogPath,
		"-g", "1",
		"-c", "2")
	if !test.IsFileHasContent(successLogPath) {
		t.Fail()
	}
	test.RemoveFile(successLogPath)

	if !test.IsFileHasContent(failLogPath) {
		t.Fail()
	}
	test.RemoveFile(failLogPath)
}

func TestAsyncFetchNoBucket(t *testing.T) {
	_, err := test.RunCmdWithError("abfetch")
	if !strings.Contains(err, "bucket can't empty") {
		t.Fail()
	}
}

func TestAsyncFetchDocument(t *testing.T) {
	test.TestDocument("abfetch", t)
}

func TestACheck(t *testing.T) {
	id := "eyJ6b25lIjoibmEwIiwicXVldWUiOiJTSVNZUEhVUy1KT0JTLVYzIiwicGFydF9pZCI6OSwib2Zmc2V0Ijo1NTEzMTU3fQ=="
	result, err := test.RunCmdWithError("acheck", test.Bucket, id)
	if len(err) > 0 || len(result) == 0{
		t.Fail()
	}
}

func TestACheckNoId(t *testing.T) {
	_, err := test.RunCmdWithError("acheck", test.Bucket)
	if !strings.Contains(err, "Id can't empty") {
		t.Fail()
	}
}

func TestACheckNoBucket(t *testing.T) {
	_, err := test.RunCmdWithError("acheck")
	if !strings.Contains(err, "Bucket can't empty") {
		t.Fail()
	}
}

func TestACheckDocument(t *testing.T) {
	test.TestDocument("acheck", t)
}
