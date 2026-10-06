package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pardnchiu/PodRun/internal/utils"
)

type PodmanArg struct {
	UID        string
	LocalDir   string
	RemoteDir  string
	Command    string
	RemoteArgs []string
	Target     string
	File       string
	Hostname   string
	IP         string

	// state
	Detach bool
}

func parseArgs(args []string) (*PodmanArg, error) {
	newArg := &PodmanArg{Target: "podman"}
	conposeExist := false

	newArg.Hostname = utils.GetHostName()

	if ip, err := utils.GetLocalIP(); err == nil {
		newArg.IP = ip
	}

	for i := 0; i < len(args); {
		arg := args[i]

		// 設定 command（第一個非參數）
		if newArg.Command == "" && !strings.HasPrefix(arg, "-") {
			newArg.Command = arg
		}

		switch {
		case arg == "-d" || arg == "--detach":
			newArg.Detach = true
			newArg.RemoteArgs = append(newArg.RemoteArgs, arg)
			i++
		case arg == "-u" && i+1 < len(args):
			newArg.UID = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "-u="):
			newArg.UID = strings.TrimPrefix(arg, "-u=")
			i++
		case strings.HasPrefix(arg, "--folder="):
			newArg.LocalDir = strings.TrimPrefix(arg, "--folder=")
			i++
		case arg == "--folder" && i+1 < len(args):
			newArg.LocalDir = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "--type="):
			newArg.Target = strings.TrimPrefix(arg, "--type=")
			i++
		case arg == "--type" && i+1 < len(args):
			newArg.Target = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "--output="):
			newArg.RemoteDir = strings.TrimPrefix(arg, "--output=")
			i++
		case arg == "--output" && i+1 < len(args):
			newArg.RemoteDir = args[i+1]
			i += 2
		case arg == "-o" && i+1 < len(args):
			newArg.RemoteDir = args[i+1]
			i += 2
		case arg == "-f" && i+1 < len(args):
			if newArg.Command == "logs" {
				// logs -f 是 follow，直接加入 RemoteArgs
				newArg.RemoteArgs = append(newArg.RemoteArgs, arg)
				i++
			} else {
				// 其他指令的 -f 是 file
				if conposeExist {
					return nil, fmt.Errorf("not supported multiple files")
				}
				conposeExist = true
				newArg.File = args[i+1]
				if newArg.LocalDir == "" {
					if dir := filepath.Dir(args[i+1]); dir != "." {
						newArg.LocalDir = dir
					}
				}
				newArg.RemoteArgs = append(newArg.RemoteArgs, "-f", filepath.Base(args[i+1]))
				i += 2
			}
		case (strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "/")) && utils.IsDir(arg):
			if newArg.LocalDir == "" {
				newArg.LocalDir = arg
			}
			i++
		default:
			newArg.RemoteArgs = append(newArg.RemoteArgs, arg)
			i++
		}
	}

	if len(newArg.RemoteArgs) > 0 {
		newArg.Command = newArg.RemoteArgs[0]
	}

	return newArg, nil
}
