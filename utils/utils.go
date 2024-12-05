package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// CheckAndInstallMpg123 检查并安装mpg123
func CheckAndInstallMpg123() {
	var cmd *exec.Cmd
	var installCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		log.Println("检测到操作系统: ", runtime.GOOS)
		cmd = exec.Command("mpg123", "--version")
		installCmd = exec.Command("sudo", "apt-get", "install", "-y", "mpg123")
	case "darwin": // macOS
		log.Println("检测到操作系统: ", runtime.GOOS)
		cmd = exec.Command("/opt/homebrew/bin/mpg123", "--version")
		installCmd = exec.Command("brew", "install", "mpg123")
	default:
		log.Println("不支持的操作系统:", runtime.GOOS)
		return
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Println("mpg123未安装，正在安装...")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		err := installCmd.Run()
		if err != nil {
			log.Fatal("安装mpg123失败:", err)
		}
		log.Println("mpg123安装成功")
	} else {
		log.Println("mpg123已安装")
	}
}

// DownloadFile 下载文件
func DownloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(out)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	_, err = io.Copy(out, resp.Body)
	return err
}

// GetLocalIP 获取本机的IP地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("无法获取本地IP地址")
}
