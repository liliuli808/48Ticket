package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	_ "gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type TicketType struct {
	Brand                string  `yaml:"brand"`
	SeatType             string  `yaml:"seatType"`
	TicketID             string  `yaml:"ticketId"`
	Cookie               string  `yaml:"cookie"`
	StartTime            string  `yaml:"startTime"`
	AddressID            string  `json:"AddressID"`
	GoodsAmount          float64 `json:"goods_amount"`
	ShippingFee          int     `json:"shipping_fee"`
	GoodsID              int     `json:"goods_id"`
	AttrID               int     `json:"attr_id"`
	Num                  string  `json:"num"`
	SumPayPrice          float64 `json:"sumPayPrice"`
	IsInv                bool    `json:"is_inv"`
	LgsID                int     `json:"lgs_id"`
	Integral             int     `json:"integral"`
	InvoiceType          int     `json:"invoice_type"`
	InvoiceTitle         string  `json:"invoice_title"`
	InvoicePrice         float64 `json:"invoice_price"`
	CompanyName          string  `json:"CompanyName"`
	TaxpayerID           string  `json:"TaxpayerID"`
	CompanyAddress       string  `json:"CompanyAddress"`
	CompanyPhone         string  `json:"CompanyPhone"`
	CompanyBankOfDeposit string  `json:"CompanyBankOfDeposit"`
	CompanyBankNo        string  `json:"CompanyBankNo"`
	RuleGoodslistContent string  `json:"rule_goodslist_content"`
	RadomRuleGoodslist   string  `json:"radom_rule_goodslist_content"`
	Remark               string  `json:"remark"`
	IsIntegralOffset     bool    `json:"IsIntegralOffsetFreight"`
	R                    float64 `json:"r"`
}

func main() {
	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone
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
	dateTime, err := time.ParseInLocation(format, ticket.StartTime, cstZone)
	if err != nil {
		fmt.Println("解析日期时间失败:", err)
		return
	}
	log.Println(dateTime.String())
	log.Println(time.Now())
	// 设置最大并发请求数
	maxConcurrentRequests := 30

	// 创建通道用于通知抢票结果
	ticketChan := make(chan bool, maxConcurrentRequests)

	// 创建等待组
	var wg sync.WaitGroup
	for {
		currentTime := time.Now().Add(2 * time.Second)
		if currentTime.Before(dateTime) {
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
			//break
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

	proxyURL, err := url.Parse("http://liliuli808:woai258123@tunnel1.docip.net:13541")
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		os.Exit(1)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	// 创建一个HTTP请求客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second,
	}

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
	req.Header.Set("Referer", "https://shop.48.cn/TOrder")
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
	if bodyMessage.HasError == false {
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
