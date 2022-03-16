package main

import (
	"fmt"
	"hostCollision/lib"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/rs/xid"
)

// 命令行选项
type Options struct {
	HostFilePath   string `short:"d" description:"host 文件路径"`
	IPFilePath     string `short:"i" description:"ip 文件路径"`
	Output         string `short:"o" description:"结果的输入路径"`
	GoRoutineLimit int    `short:"n" default:"20" description:"协程的并发数量"`
	SleepTime      int    `short:"s" default:"1000" description:"sleep (ms)"`
	MaxHost        int    `short:"m" default:"50" description:"允许爆破的最大的MaxHost"`
	Rate           int    `short:"r" default:"85" description:"两个页面的相似度"`
}

func init() {
	// 日志文件夹
	logDirPath := filepath.Join(".", "log")
	err := os.MkdirAll(logDirPath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// 日志文件
	logPath := fmt.Sprintf("%s/log_%d%s", logDirPath, time.Now().Unix(), ".txt")
	logFile, err := os.Create(logPath)
	if err != nil {
		panic(err)
	}

	log.SetOutput(logFile)
	log.SetPrefix("[HostCollision]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

}

func main() {

	//
	// 解析命令行参数
	//
	var opt Options
	p := flags.NewParser(&opt, flags.Default)
	_, err := p.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	//
	// 打印banner
	//
	lib.Banner()

	//
	// 打印当前的一些基础配置
	//
	log.Printf("---------------------------------")
	log.Printf("SleepTime: %d ms", opt.SleepTime)
	log.Printf("MaxHost: %d", opt.MaxHost)
	log.Printf("simRate: %d", opt.Rate)
	log.Printf("GoRoutine: %d", opt.GoRoutineLimit)
	log.Printf("---------------------------------")

	//
	// 读取输入文件
	//
	hosts, err := lib.ReadFile(opt.HostFilePath)
	if err != nil {
		log.Println(err)
		return
	}

	ips, err := lib.ReadFile(opt.IPFilePath)
	if err != nil {
		log.Println(err)
		return
	}

	//
	// 开始 host 爆破
	//

	// 限制协程的数量
	chLimit := make(chan int, opt.GoRoutineLimit)

	var wg sync.WaitGroup

	// 保存结果的文件
	resultFile, err := os.OpenFile(opt.Output, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}
	defer resultFile.Close()

	for _, protocol := range []string{"http", "https"} {
		for _, ip := range ips {

			// 存储结果
			chResult := make(chan string, 200)

			// 纯 ip 去访问
			resp1, err := lib.SendRequest(protocol, ip, "")
			if err != nil {
				log.Println(err)
				continue
			}

			// 带上一个不存在的Host去访问
			hostNotExist := fmt.Sprintf("%s.%s", xid.New().String(), "aaa.com")
			resp2, err := lib.SendRequest(protocol, ip, hostNotExist)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, host := range hosts {
				wg.Add(1)
				chLimit <- 1
				fmt.Printf("开始测试:%s://%s, host: %s\n", protocol, ip, host)
				go func(protocolTemp, ipTemp, hostTemp string, resp1Temp, resp2Temp lib.Resp, waitGroup *sync.WaitGroup) {
					defer waitGroup.Done()

					// 带上host去访问
					resp3, err := lib.SendRequest(protocolTemp, ipTemp, hostTemp)
					if err != nil {
						log.Println(err)
						<-chLimit
						return
					}

					// 对比1，3 的结果
					sim1 := lib.ContentSim(resp1Temp.Content, resp2Temp.Content)

					// 对比2，3 的结果
					sim2 := lib.ContentSim(resp2Temp.Content, resp3.Content)

					log.Printf("statusCode: %3d | protocol: %5s | ip: %16s | host: %-30s | simRate1: %3d | simRate2:%3d | contentLength: %6d", resp3.StatusCode, protocolTemp, ipTemp, hostTemp, sim1, sim2, resp3.ContentLength)
					if resp3.StatusCode == 200 && sim1 < opt.Rate && sim2 < opt.Rate {
						result := fmt.Sprintf("%s,%s,%s\n", protocolTemp, ipTemp, hostTemp)
						fmt.Printf("发现：%s", result)
						chResult <- result
					}
					<-chLimit
				}(protocol, ip, host, resp1, resp2, &wg)
			}
			wg.Wait()
			close(chResult)

			if len(chResult) > opt.MaxHost {
				log.Printf("%s://%s 超过%d次， 请确认！\n", protocol, ip, opt.MaxHost)
			} else {
				for res := range chResult {
					resultFile.WriteString(res)
				}
			}
		}
	}
	close(chLimit)

}
