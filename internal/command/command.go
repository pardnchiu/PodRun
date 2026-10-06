package command

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pardnchiu/PodRun/internal/utils"
)

func New() (*PodmanArg, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("[x] podrun <command> [args...]")
	}

	command := os.Args[1]
	switch command {
	case "info":
		fmt.Println("show project info")
	case "export":
		fmt.Println("export project to pod manifest")
	case "deploy":
		fmt.Println("deploy project to kubernetes")
	case "clone":
		fmt.Println("clone project to local")
	case "domain":
		fmt.Println("set domain to pod")
	}

	args, err := parseArgs(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("[x] %v", err)
	}

	if len(args.RemoteArgs) == 0 {
		return nil, fmt.Errorf("[x] please ensure docker compose <command> [args...] is valid first before running podrun")
	}

	localDir, err := getLocalDir(args.LocalDir)
	if err != nil {
		return nil, fmt.Errorf("[x] %v", err)
	}
	args.LocalDir = localDir

	uid, remoteDir, err := setRemoteDir(localDir)
	if err != nil {
		return nil, fmt.Errorf("[x] %v", err)
	}
	args.RemoteDir = remoteDir

	if args.UID == "" {
		args.UID = uid
	}

	return args, nil
}

func getLocalDir(folder string) (string, error) {
	var err error
	newFolder := folder
	if newFolder == "" {
		newFolder, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	absPath, err := filepath.Abs(newFolder)
	if err != nil {
		return "", err
	}
	if !utils.IsDir(absPath) {
		return "", fmt.Errorf("folder does not exist: %s", absPath)
	}
	if !utils.FileExist(filepath.Join(absPath, "docker-compose.yml")) &&
		!utils.FileExist(filepath.Join(absPath, "docker-compose.yaml")) {
		return "", fmt.Errorf("docker-compose.yml not found in folder: %s", absPath)
	}
	return absPath, nil
}

func setRemoteDir(localFolder string) (string, string, error) {
	mac, err := utils.GetMAC()
	if err != nil {
		mac, _ = os.Hostname()
	}
	hash := md5.Sum(fmt.Appendf(nil, "%s@%s", mac, localFolder))
	return hex.EncodeToString(hash[:]),
		filepath.Join("/home/podrun",
			fmt.Sprintf("%s_%s", filepath.Base(localFolder), hex.EncodeToString(hash[:])[:8])),
		nil
}
