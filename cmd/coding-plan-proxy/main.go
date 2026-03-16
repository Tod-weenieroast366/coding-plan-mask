// Coding Plan Proxy - 本地代理转发工具
// 将请求转发到云厂商 Coding Plan API
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"coding-plan-proxy/internal/config"
	"coding-plan-proxy/internal/server"
	"coding-plan-proxy/internal/storage"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version = "2.0.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// 检查子命令
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "show", "info", "connection":
			showConnection(os.Args[2:])
			return
		case "stats":
			showStats(os.Args[2:])
			return
		case "monitor":
			showMonitor(os.Args[2:])
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	// 命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	provider := flag.String("provider", "", "服务商名称")
	apiKey := flag.String("api-key", "", "Coding Plan API Key")
	localAPIKey := flag.String("local-api-key", "", "本地 API Key")
	host := flag.String("host", "", "监听地址")
	port := flag.Int("port", 0, "监听端口")
	debug := flag.Bool("debug", false, "调试模式")
	general := flag.Bool("general", false, "使用通用 API 端点")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Coding Plan Proxy %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 命令行参数覆盖
	if *provider != "" {
		cfg.Provider = *provider
	}
	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}
	if *localAPIKey != "" {
		cfg.LocalAPIKey = *localAPIKey
	}
	if *host != "" {
		cfg.ListenHost = *host
	}
	if *port != 0 {
		cfg.ListenPort = *port
	}
	if *debug {
		cfg.Debug = true
	}
	if *general {
		cfg.UseCodingEndpoint = false
	}

	// 初始化日志
	logger := initLogger(cfg.Debug)
	defer logger.Sync()

	// 初始化存储
	dataDir := filepath.Join(filepath.Dir(cfg.GetConfigPath()), "data")
	store, err := storage.New(dataDir)
	if err != nil {
		logger.Fatal("初始化存储失败", zap.Error(err))
	}

	// 打印启动信息
	printBanner(cfg, logger)

	// 检查必要配置
	if cfg.APIKey == "" {
		logger.Warn("未配置 Coding Plan API Key，请使用 --api-key 参数或配置文件设置")
	}

	if cfg.LocalAPIKey == "" {
		logger.Warn("未配置本地 API Key，代理将允许任意客户端连接（不推荐）")
	}

	// 创建并启动服务器
	srv := server.New(cfg, logger, store)
	if err := srv.Start(); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}

// showStats 显示统计信息
func showStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	configPath := fs.String("config", "", "配置文件路径")
	_ = fs.Parse(args)

	// 确定数据目录
	var dataDir string
	if *configPath != "" {
		// 从配置文件路径推导数据目录
		dataDir = filepath.Join(filepath.Dir(*configPath), "data")
	} else {
		// 尝试从 systemd 服务获取配置路径
		serviceConfig := "/opt/project/coding-plan-proxy/config/config.toml"
		if _, err := os.Stat(serviceConfig); err == nil {
			dataDir = filepath.Join(filepath.Dir(serviceConfig), "data")
		} else {
			// 默认路径
			homeDir, _ := os.UserHomeDir()
			dataDir = filepath.Join(homeDir, ".config", "coding-plan-proxy", "data")
		}
	}
	dbPath := filepath.Join(dataDir, "proxy.db")

	// 检查数据库是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("统计数据库不存在，服务可能还未运行过")
		return
	}

	// 打开存储
	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开存储失败: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// 获取统计
	stats, err := store.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取统计失败: %v\n", err)
		os.Exit(1)
	}

	// 输出
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Token 使用统计                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  总请求数:     %-42d ║\n", stats.TotalRequests)
	fmt.Printf("║  总上传 Token: %-42d ║\n", stats.TotalInputTokens)
	fmt.Printf("║  总下载 Token: %-42d ║\n", stats.TotalOutputTokens)
	fmt.Printf("║  总 Token:     %-42d ║\n", stats.TotalTokens)
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  今日请求:     %-42d ║\n", stats.TodayRequests)
	fmt.Printf("║  今日上传:     %-42d ║\n", stats.TodayInputTokens)
	fmt.Printf("║  今日下载:     %-42d ║\n", stats.TodayOutputTokens)
	fmt.Printf("║  今日 Token:   %-42d ║\n", stats.TodayTokens)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// showMonitor 实时监控（带图表）
func showMonitor(args []string) {
	fs := flag.NewFlagSet("monitor", flag.ExitOnError)
	configPath := fs.String("config", "", "配置文件路径")
	interval := fs.Int("interval", 2, "刷新间隔(秒)")
	_ = fs.Parse(args)

	// 确定数据目录
	var dataDir string
	if *configPath != "" {
		dataDir = filepath.Join(filepath.Dir(*configPath), "data")
	} else {
		serviceConfig := "/opt/project/coding-plan-proxy/config/config.toml"
		if _, err := os.Stat(serviceConfig); err == nil {
			dataDir = filepath.Join(filepath.Dir(serviceConfig), "data")
		} else {
			homeDir, _ := os.UserHomeDir()
			dataDir = filepath.Join(homeDir, ".config", "coding-plan-proxy", "data")
		}
	}

	// 历史数据（用于图表）
	var historyReqs, historyIn, historyOut []int64
	maxHistory := 30

	// 清屏并隐藏光标
	fmt.Print("\033[2J\033[?25l")

	// 捕获 Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		// 恢复光标
		fmt.Print("\033[?25h")
		fmt.Println("\n监控已停止")
	}()

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	// 首次立即显示
	displayMonitor(dataDir, &historyReqs, &historyIn, &historyOut, maxHistory)

	for {
		select {
		case <-sigChan:
			return
		case <-ticker.C:
			displayMonitor(dataDir, &historyReqs, &historyIn, &historyOut, maxHistory)
		}
	}
}

// displayMonitor 显示监控界面
func displayMonitor(dataDir string, historyReqs, historyIn, historyOut *[]int64, maxHistory int) {
	// 打开存储获取实时数据
	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Printf("打开存储失败: %v\n", err)
		return
	}

	stats, err := store.GetStats()
	store.Close()

	if err != nil {
		fmt.Printf("获取统计失败: %v\n", err)
		return
	}

	// 更新历史数据
	*historyReqs = append(*historyReqs, stats.TodayRequests)
	*historyIn = append(*historyIn, stats.TodayInputTokens)
	*historyOut = append(*historyOut, stats.TodayOutputTokens)

	if len(*historyReqs) > maxHistory {
		*historyReqs = (*historyReqs)[1:]
		*historyIn = (*historyIn)[1:]
		*historyOut = (*historyOut)[1:]
	}

	// 计算速率
	var rateReqs, rateIn, rateOut float64
	if len(*historyReqs) >= 2 {
		rateReqs = float64((*historyReqs)[len(*historyReqs)-1] - (*historyReqs)[len(*historyReqs)-2])
		rateIn = float64((*historyIn)[len(*historyIn)-1] - (*historyIn)[len(*historyIn)-2])
		rateOut = float64((*historyOut)[len(*historyOut)-1] - (*historyOut)[len(*historyOut)-2])
	}

	// 移动光标到屏幕开头
	fmt.Print("\033[H")

	// 标题
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\033[36m╔══════════════════════════════════════════════════════════════════════╗\033[0m\n")
	fmt.Printf("\033[36m║\033[0m  \033[1;37mCoding Plan Proxy 实时监控\033[0m                              \033[90m%s\033[0m  \033[36m║\033[0m\n", now)
	fmt.Printf("\033[36m╠══════════════════════════════════════════════════════════════════════╣\033[0m\n")

	// 统计面板
	fmt.Printf("\033[36m║\033[0m  \033[33m总计\033[0m                                                              \033[36m║\033[0m\n")
	fmt.Printf("\033[36m║\033[0m    请求数: \033[32m%-10d\033[0m  上传Token: \033[32m%-10d\033[0m  下载Token: \033[32m%-10d\033[0m  \033[36m║\033[0m\n",
		stats.TotalRequests, stats.TotalInputTokens, stats.TotalOutputTokens)
	fmt.Printf("\033[36m╠══════════════════════════════════════════════════════════════════════╣\033[0m\n")

	// 今日统计
	fmt.Printf("\033[36m║\033[0m  \033[33m今日\033[0m                                                              \033[36m║\033[0m\n")
	fmt.Printf("\033[36m║\033[0m    请求数: \033[32m%-10d\033[0m  上传Token: \033[32m%-10d\033[0m  下载Token: \033[32m%-10d\033[0m  \033[36m║\033[0m\n",
		stats.TodayRequests, stats.TodayInputTokens, stats.TodayOutputTokens)
	fmt.Printf("\033[36m╠══════════════════════════════════════════════════════════════════════╣\033[0m\n")

	// 速率
	fmt.Printf("\033[36m║\033[0m  \033[33m实时速率 (每周期)\033[0m                                               \033[36m║\033[0m\n")
	fmt.Printf("\033[36m║\033[0m    请求: \033[35m%-8.0f\033[0m  上传: \033[35m%-10.0f\033[0m  下载: \033[35m%-10.0f\033[0m       \033[36m║\033[0m\n",
		rateReqs, rateIn, rateOut)
	fmt.Printf("\033[36m╠══════════════════════════════════════════════════════════════════════╣\033[0m\n")

	// ASCII 图表 - 请求趋势
	fmt.Printf("\033[36m║\033[0m  \033[33m请求趋势图\033[0m                                                          \033[36m║\033[0m\n")
	drawChart(*historyReqs, 8, 60, "reqs")
	fmt.Printf("\033[36m╠══════════════════════════════════════════════════════════════════════╣\033[0m\n")

	// ASCII 图表 - Token 趋势
	fmt.Printf("\033[36m║\033[0m  \033[33mToken 趋势图 (上传/下载)\033[0m                                          \033[36m║\033[0m\n")
	drawDualChart(*historyIn, *historyOut, 4, 60)
	fmt.Printf("\033[36m╚══════════════════════════════════════════════════════════════════════╝\033[0m\n")

	fmt.Printf("\033[90m按 Ctrl+C 退出\033[0m")
}

// drawChart 绘制单线图表
func drawChart(data []int64, height, width int, label string) {
	if len(data) == 0 {
		for i := 0; i < height; i++ {
			fmt.Printf("\033[36m║\033[0m  \033[90m%-60s\033[0m  \033[36m║\033[0m\n", "")
		}
		return
	}

	// 找最大值
	maxVal := int64(1)
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}

	// 绘制图表
	for row := height - 1; row >= 0; row-- {
		fmt.Printf("\033[36m║\033[0m  ")
		threshold := int64(row) * maxVal / int64(height)
		for col := 0; col < width; col++ {
			dataIdx := col * len(data) / width
			if dataIdx >= len(data) {
				dataIdx = len(data) - 1
			}
			if data[dataIdx] >= threshold {
				fmt.Printf("\033[32m█\033[0m")
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Printf(" \033[36m║\033[0m\n")
	}
}

// drawDualChart 绘制双线图表
func drawDualChart(data1, data2 []int64, height, width int) {
	if len(data1) == 0 && len(data2) == 0 {
		for i := 0; i < height; i++ {
			fmt.Printf("\033[36m║\033[0m  \033[90m%-60s\033[0m  \033[36m║\033[0m\n", "")
		}
		return
	}

	// 找最大值
	maxVal := int64(1)
	for _, v := range data1 {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range data2 {
		if v > maxVal {
			maxVal = v
		}
	}

	// 绘制图表
	for row := height - 1; row >= 0; row-- {
		fmt.Printf("\033[36m║\033[0m  ")
		threshold := int64(row) * maxVal / int64(height)
		for col := 0; col < width; col++ {
			dataIdx := col * max(len(data1), len(data2)) / width
			var has1, has2 bool
			if dataIdx < len(data1) {
				has1 = data1[dataIdx] >= threshold
			}
			if dataIdx < len(data2) {
				has2 = data2[dataIdx] >= threshold
			}

			if has1 && has2 {
				fmt.Printf("\033[35m█\033[0m") // 紫色表示重叠
			} else if has1 {
				fmt.Printf("\033[32m█\033[0m") // 绿色表示上传
			} else if has2 {
				fmt.Printf("\033[34m█\033[0m") // 蓝色表示下载
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Printf(" \033[36m║\033[0m \033[32m↑\033[0m上传 \033[34m↓\033[0m下载\n")
	}
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// showConnection 显示连接信息
func showConnection(args []string) {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	configPath := fs.String("config", "", "配置文件路径")
	jsonOutput := fs.Bool("json", false, "JSON 格式输出")
	_ = fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	baseURL := fmt.Sprintf("http://%s:%d/v1", cfg.ListenHost, cfg.ListenPort)

	if *jsonOutput {
		output := map[string]string{
			"base_url": baseURL,
			"api_key":  cfg.LocalAPIKey,
		}
		json.NewEncoder(os.Stdout).Encode(output)
	} else {
		fmt.Println()
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║              本地连接信息 (Local Connection)                ║")
		fmt.Println("╠════════════════════════════════════════════════════════════╣")
		fmt.Printf("║  Base URL:  %-45s ║\n", baseURL)
		if cfg.LocalAPIKey != "" {
			fmt.Printf("║  API Key:   %-45s ║\n", cfg.LocalAPIKey)
		} else {
			fmt.Printf("║  API Key:   %-45s ║\n", "(未设置，无需认证)")
		}
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("客户端配置示例:")
		fmt.Println("```json")
		if cfg.LocalAPIKey != "" {
			fmt.Printf(`{
    "base_url": "%s",
    "api_key": "%s",
    "model": "glm-4-flash"
}`, baseURL, cfg.LocalAPIKey)
		} else {
			fmt.Printf(`{
    "base_url": "%s",
    "model": "glm-4-flash"
}`, baseURL)
		}
		fmt.Println("\n```")
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Printf(`Coding Plan Proxy v%s - 本地代理转发工具

用法:
  %s [选项]           启动代理服务
  %s show             显示本地连接信息
  %s show --json      JSON 格式输出连接信息
  %s stats            显示 Token 使用统计
  %s monitor          实时监控（带图表）
  %s monitor -interval 5  设置刷新间隔为5秒

子命令:
  show, info, connection    显示本地连接地址和 API Key
  stats                      显示 Token 使用统计
  monitor                    实时监控（带 ASCII 图表）

选项:
  -config string         配置文件路径
  -provider string       服务商 (zhipu, zhipu_v2, aliyun, minimax, deepseek, moonshot)
  -api-key string        Coding Plan API Key
  -local-api-key string  本地 API Key
  -host string           监听地址 (默认 127.0.0.1)
  -port int              监听端口 (默认 8787)
  -debug                 调试模式
  -general               使用通用 API 端点
  -version               显示版本信息

伪装工具配置 (在 config.toml 中设置):
  disguise_tool = "opencode"   伪装为 OpenCode (默认)
  disguise_tool = "openclaw"   伪装为 OpenClaw
  disguise_tool = "custom"     使用自定义 User-Agent

示例:
  # 启动服务
  %s -api-key sk-xxx -local-api-key sk-local-xxx

  # 显示连接信息
  %s show

  # 显示统计
  %s stats

  # 实时监控
  %s monitor
`, version, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

// initLogger 初始化日志
func initLogger(debug bool) *zap.Logger {
	var zcfg zap.Config
	if debug {
		zcfg = zap.NewDevelopmentConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zcfg = zap.NewProductionConfig()
		zcfg.EncoderConfig.TimeKey = "time"
		zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := zcfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	return logger
}

// printBanner 打印启动横幅
func printBanner(cfg *config.Config, logger *zap.Logger) {
	provider, err := cfg.GetProviderConfig()
	providerName := "未知"
	if err == nil {
		providerName = provider.Name
	}

	endpointType := "Coding Plan"
	if !cfg.UseCodingEndpoint {
		endpointType = "通用 API"
	}

	localAuth := "已配置"
	if cfg.LocalAPIKey == "" {
		localAuth = "未配置 (公开模式)"
	}

	apiKeyStatus := "已配置"
	if cfg.APIKey == "" {
		apiKeyStatus = "未配置"
	}

	debugMode := "关闭"
	if cfg.Debug {
		debugMode = "开启"
	}

	// 获取伪装工具信息
	disguiseTool := cfg.DisguiseTool
	if disguiseTool == "" {
		disguiseTool = "opencode"
	}
	toolInfo, ok := config.PredefinedDisguiseTools[disguiseTool]
	toolName := "未知"
	if ok {
		toolName = toolInfo.Name
	}
	userAgent := cfg.GetEffectiveUserAgent()

	banner := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║                Coding Plan Proxy v%s                      ║
╠══════════════════════════════════════════════════════════════╣
║  服务商: %-50s ║
║  端点类型: %-48s ║
║  监听地址: http://%s:%-39d ║
║  本地认证: %-48s ║
║  Coding Key: %-46s ║
║  伪装工具: %-48s ║
║  User-Agent: %-46s ║
║  调试模式: %-48s ║
╚══════════════════════════════════════════════════════════════╝
`, version, padRight(providerName, 50), padRight(endpointType, 48),
		cfg.ListenHost, cfg.ListenPort,
		padRight(localAuth, 48), padRight(apiKeyStatus, 46),
		padRight(toolName, 48), padRight(userAgent, 46), padRight(debugMode, 48))

	fmt.Print(banner)

	logger.Info("服务启动",
		zap.String("provider", cfg.Provider),
		zap.String("listen", fmt.Sprintf("%s:%d", cfg.ListenHost, cfg.ListenPort)),
		zap.String("disguise", disguiseTool),
	)
}

// padRight 右侧填充
func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}
