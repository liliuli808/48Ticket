package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type TicketType struct {
	Brand     string `yaml:"brand"`
	SeatType  string `yaml:"seatType"`
	TicketID  string `yaml:"ticketId"`
	Cookie    string `yaml:"cookie"`
	StartTime string `yaml:"startTime"`
	Num       string `yaml:"num"`
}

func main() {
	// 读取YAML配置文件
	yamlFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("无法读取YAML文件: %v", err)
	}

	// 创建TicketType结构体实例
	var ticket TicketType

	// 解析YAML配置文件到结构体
	err = yaml.Unmarshal(yamlFile, &ticket)
	if err != nil {
		log.Fatalf("无法解析YAML文件: %v", err)
	}

	// 定义日期时间字符串的格式
	format := "2006-01-02 15:04:05"

	// 使用 time 包中的 Parse 函数将字符串解析为时间对象
	dateTime, err := time.ParseInLocation(format, ticket.StartTime, time.Local)
	if err != nil {
		fmt.Println("解析日期时间失败:", err)
		return
	}

	// 设置最大并发请求数
	maxConcurrentRequests := 2

	// 创建通道用于通知抢票结果
	ticketChan := make(chan bool, 2)

	// 创建等待组
	var wg sync.WaitGroup
	for {
		currentTime := time.Now().Add(10 * time.Second)
		if currentTime.Before(dateTime.In(time.Local)) {
			continue
		}
		// 启动最大并发请求数的抢票任务
		for i := 0; i < maxConcurrentRequests; i++ {
			wg.Add(1)
			go ticketAdd(ticket, ticketChan, &wg)
		}

		// 等待所有抢票任务完成
		wg.Wait()

		// 统计抢票结果
		successCount := 0
		for i := 0; i < maxConcurrentRequests; i++ {
			if <-ticketChan {
				successCount++
			}
		}

		if successCount > 0 {
			log.Println("抢票成功")
			break
		} else {
			log.Println("抢票失败，继续尝试...")
		}
	}
	close(ticketChan)
}

func ticketAdd(ticket TicketType, ticketChan chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	requestData := "ticketsid=%s&num=%s&seattype=%s&brand_id=%s&choose_times_end=-1&ticketstype=2&r=0.056981472084815854"
	requestURL := "https://shop.48.cn/TOrder/ticket_Add"

	requestData = fmt.Sprintf(requestData, ticket.TicketID, ticket.Num, ticket.SeatType, ticket.Brand)

	// 创建一个HTTP请求客户端
	client := &http.Client{}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(requestData))
	if err != nil {
		log.Println(err)
		ticketChan <- false
		return
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", ticket.Cookie)
	req.Header.Set("Origin", "https://shop.48.cn")
	req.Header.Set("Referer", "https://shop.48.cn/tickets/item/5421?seat_type=4")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Linux"`)

	// 发送HTTP请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		ticketChan <- false
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		ticketChan <- false
		return
	}
	var bodyMessage Message
	err = json.Unmarshal(body, &bodyMessage)
	if err != nil {
		log.Println(err)
		ticketChan <- false
		return
	}
	log.Println(bodyMessage)
	if bodyMessage.ErrorCode == "144006" || bodyMessage.ErrorCode == "144008" {
		ticketChan <- true
		return
	}

	ticketChan <- false
	return
}

type Message struct {
	HasError     bool   `json:"HasError"`
	ErrorCode    string `json:"ErrorCode"`
	Message      string `json:"Message"`
	ReturnObject string `json:"ReturnObject"`
}
