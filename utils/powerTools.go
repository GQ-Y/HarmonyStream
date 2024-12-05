package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var isConnected = false // 连接状态

func StartApplication() {
	logFilePath := "app.log"
	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatal("无法删除旧的日志文件：", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件：", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	CheckAndInstallMpg123()
	clearAudioCache()

	if err = GetAndSaveMACAddressToTXT(); err != nil {
		log.Fatal("获取和保存唯一地址失败:", err)
	}

	u := url.URL{
		Scheme: "wss",
		Host:   "cd.api.yingzhu.net",
		Path:   "/screen.io",
	}

	log.Println("正在连接到", u.String())
	fmt.Println("启动成功，程序正在运行中")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Println("连接失败:", err)
			isConnected = false
			time.Sleep(5 * time.Second) // 连接失败后重试
			continue
		}

		isConnected = true // 连接成功
		defer func() {
			if err := c.Close(); err != nil {
				log.Println("关闭连接时出错:", err)
			}
		}()

		macAddress := GetMACAddress()
		if macAddress == "" {
			log.Fatal("获取地址失败")
		}

		localIP, err := GetLocalIP()
		if err != nil {
			log.Println("获取本地 IP 地址失败:", err)
			return
		}

		// 启动心跳机制
		go startHeartbeat(c, localIP, macAddress)

		// 持续读取消息
		readMessages(c)

		// 如果 readMessages 返回，意味着连接已关闭，需要重连
		log.Println("重连中...")
		isConnected = false
		time.Sleep(5 * time.Second) // 等待一段时间再尝试重连
	}
}

func GetMACAddress() string {
	macFilePath := "cache.txt"
	macAddress, err := os.ReadFile(macFilePath)
	if err == nil {
		return strings.TrimSpace(string(macAddress))
	}

	// 如果文件不存在，生成新的地址
	macAddress = []byte(generateUniqueID())

	if err := os.WriteFile(macFilePath, macAddress, 0666); err != nil {
		log.Println("写入文件失败:", err)
		return ""
	}

	log.Println("成功写入地址到文件:", string(macAddress)) // 添加调试信息
	return string(macAddress)
}

func clearAudioCache() {
	cacheDir := "audio_cache"
	if err := os.RemoveAll(cacheDir); err != nil {
		log.Println("无法清理音频缓存目录:", err)
	} else {
		log.Println("音频缓存已清理")
	}
}

func sendHeartbeat(c *websocket.Conn, localIP, macAddress string) {
	if !isConnected || c == nil { // 如果未连接或连接为空，不发送心跳
		return
	}

	heartbeat := map[string]interface{}{
		"event": "device",
		"data": map[string]interface{}{
			"app_version": "V1.0.0.0",
			"device_ip":   localIP,
			"device_name": "智能功放",
			"brand_name":  "赢筑",
			"device_type": 90014,
			"mac_address": macAddress,
			"online":      1,
		},
	}

	heartbeatJSON, err := json.Marshal(heartbeat)
	if err != nil {
		log.Println("JSON 序列化错误:", err)
		return
	}

	// 重试发送心跳消息
	for retry := 0; retry < 3; retry++ {
		if err = c.WriteMessage(websocket.TextMessage, heartbeatJSON); err == nil {
			log.Printf("设备上线通知: %s", heartbeatJSON)
			return
		}
		log.Println("写入消息错误:", err)
		time.Sleep(2 * time.Second) // 等待一段时间后重试
	}
	log.Println("发送心跳消息失败，已达最大重试次数")
}

func sendHeartbeatEmpty(c *websocket.Conn) {
	if !isConnected || c == nil { // 如果未连接或连接为空，不发送心跳
		return
	}
	if err := c.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		log.Println("写入心跳消息错误:", err)
	} else {
		log.Printf("ping Server 成功")
	}
}

func startHeartbeat(c *websocket.Conn, localIP, macAddress string) {
	ticker := time.NewTicker(30 * time.Second) // 每 30 秒发送一次心跳
	defer ticker.Stop()

	sendHeartbeat(c, localIP, macAddress) // 初始发送设备状态

	for range ticker.C {
		sendHeartbeatEmpty(c) // 发送Ping消息
	}
}

func readMessages(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("读取消息错误:", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("意外关闭错误，正在重连...")
				isConnected = false
				return
			}
			return
		}

		log.Printf("接收到: %s", message)

		messageStr := string(message)
		log.Printf("处理后的消息: %s", messageStr)

		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(messageStr), &msg); err != nil {
			log.Println("JSON 反序列化错误:", err)
			continue
		}

		if data, ok := msg["data"].(map[string]interface{}); ok {
			if typ, ok := data["type"].(string); ok && typ == "play" {
				if audioURL, ok := data["audio"].(string); ok {
					playAudio(audioURL)
				}
			}
		}
	}
}

func playAudio(audioURL string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("获取当前工作目录失败:", err)
		return
	}

	cacheDir := filepath.Join(wd, "audio_cache")
	fileName := filepath.Join(cacheDir, filepath.Base(audioURL))

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Println("创建音频缓存目录失败:", err)
		return
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if err := DownloadFile(audioURL, fileName); err != nil {
			log.Println("下载音频文件失败:", err)
			return
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("mpg123"); err != nil {
			log.Println("mpg123 未安装")
			return
		}
		cmd = exec.Command("mpg123", fileName)
	case "darwin":
		if _, err := exec.LookPath("afplay"); err != nil {
			log.Println("afplay 未安装")
			return
		}
		cmd = exec.Command("afplay", fileName)
	case "windows":
		cmd = exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer '"+fileName+"').PlaySync();")
	default:
		log.Println("不支持的操作系统:", runtime.GOOS)
		return
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Println("播放音频文件失败:", err)
		log.Println("详细错误信息:", out.String())
	}
	log.Printf("尝试播放音频来自: %s", fileName)
}

func GetAndSaveMACAddressToTXT() error {
	macFilePath := "cache.txt"
	if macAddress, err := os.ReadFile(macFilePath); err == nil {
		log.Println("唯一设备码已存在:", string(macAddress))
		return nil // 直接返回，不重新生成
	}

	// 如果文件不存在，生成新的地址
	macAddress := generateUniqueID()
	log.Println("生成新的地址:", macAddress) // 添加调试信息

	if err := os.WriteFile(macFilePath, []byte(macAddress), 0666); err != nil {
		log.Println("写入文件失败:", err) // 添加调试信息
		return err
	}

	log.Println("成功写入地址到文件:", macAddress) // 添加调试信息
	return nil
}

// 生成唯一标识的函数
func generateUniqueID() string {
	rand.Seed(time.Now().UnixNano())                               // 确保每次运行都有不同的随机数
	randomNumber := fmt.Sprintf("%012d", rand.Intn(1000000000000)) // 生成 12 位字符串
	timestamp := time.Now().Unix()                                 // 获取当前时间戳
	return fmt.Sprintf("%s%d", randomNumber, timestamp)            // 返回拼接后的唯一标识
}
