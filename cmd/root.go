package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/qiniu/go-sdk/v7/client"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/qiniu/qshell/v2/iqshell/common/log"
	"github.com/qiniu/qshell/v2/iqshell/common/workspace"
	"github.com/spf13/cobra"
)

var (
	// 开启命令行的调试模式
	DebugFlag     bool
	DeepDebugInfo bool

	// qshell 版本信息， qshell -v
	VersionFlag bool
	cfgFile     string
	local       bool
)

const (
	bash_completion_func = `__qshell_parse_get()
{
    local qshell_output out
    if qshell_output=$(qshell user ls --name 2>/dev/null); then
        out=($(echo "${qshell_output}"))
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
    fi
}

__qshell_get_resource()
{
    __qshell_parse_get
    if [[ $? -eq 0 ]]; then
        return 0
    fi
}

__custom_func() {
    case ${last_command} in
        qshell_user_cu)
            __qshell_get_resource
            return
            ;;
        *)
            ;;
    esac
}
`
)

// cobra root cmd, all other commands is children or subchildren of this root cmd
var RootCmd = &cobra.Command{
	Use:                    "qshell",
	Short:                  "Qiniu commandline tool for managing your bucket and CDN",
	Version:                version,
	BashCompletionFunction: bash_completion_func,
}

var initFuncs []func()

func OnInitialize(f ...func()) {
	initFuncs = append(initFuncs, f...)
}

func init() {
	cobra.OnInitialize(func() {
		initConfig()
		for _, f := range initFuncs {
			f()
		}
	})

	RootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "debug mode")
	RootCmd.PersistentFlags().BoolVarP(&DeepDebugInfo, "ddebug", "D", false, "deep debug mode")
	RootCmd.PersistentFlags().BoolVarP(&VersionFlag, "version", "v", false, "show version")
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "C", "", "config file (default is $HOME/.qshell.json)")
	RootCmd.PersistentFlags().BoolVarP(&local, "local", "L", false, "use current directory as config file path")
}

func initConfig() {
	//set cpu count
	runtime.GOMAXPROCS(runtime.NumCPU())
	//set qshell user agent
	storage.UserAgent = UserAgent()

	// 加载 log
	logLevel := log.LevelInfo
	if DebugFlag {
		logLevel = log.LevelDebug
	}
	if DeepDebugInfo {
		logLevel = log.LevelVerbose
		client.TurnOnDebug()
	}
	log.LoadConsole(logLevel)

	// 加载工作区
	workspacePath := ""
	if local {
		dir, gErr := os.Getwd()
		if gErr != nil {
			fmt.Fprintf(os.Stderr, "get current directory: %v\n", gErr)
			os.Exit(1)
		}
		workspacePath = dir
	}

	workspace.Load(workspace.Workspace(workspacePath), workspace.UserConfigPath(cfgFile))
}
