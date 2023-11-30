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
	"os"
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
	LogFile              string  `json:"logFile"`
}

func main() {

	// 创建或打开一个日志文件
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer logFile.Close()

	// 设置日志输出到文件
	log.SetOutput(logFile)

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

	currentTime := time.Now().Add(2 * time.Second)
	if currentTime.Before(dateTime) {
		return
	}

	go func() {
		for {
			ticketAdd(ticket)
		}
	}()

	checkResultChan := make(chan bool)
	go func() {
		for {

			if ticketCheck(ticket) {
				checkResultChan <- true
				break
			}

			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case result := <-checkResultChan:
		if result {
			log.Println("抢票成功")
			break
		}
	case <-time.After(2 * time.Minute):
		break
	}

}

func ticketCheck(ticket TicketType) bool {
	checkUrl := "https://shop.48.cn/TOrder/tickCheck?id=%s&seattype=%s&r=0.5246474955150733"
	requestUrl := fmt.Sprintf(checkUrl, ticket.TicketID, ticket.SeatType)
	log.Print(requestUrl, "\n")
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new HTTP request with the specified method and URL
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}
	// Set the request headers
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", ticket.Cookie)
	req.Header.Set("Referer", "https://shop.48.cn/tickets/item/5633?seat_type=3")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	var checkMessage CheckMessage
	err = json.Unmarshal(body, &checkMessage)
	if err != nil {
		log.Println(err)
	}
	log.Println(checkMessage)
	if checkMessage.HasError == true {
		return false
	}
	if checkMessage.ErrorCode == "wait" {
		return false
	}
	return true
}

type CheckMessage struct {
	HasError     bool        `json:"HasError"`
	ErrorCode    string      `json:"ErrorCode"`
	Message      interface{} `json:"Message"`
	ReturnObject string      `json:"ReturnObject"`
}

func ticketAdd(ticket TicketType) {
	requestData := "ticketsid=%s&num=%s&seattype=%s&brand_id=%s&choose_times_end=-1&ticketstype=2&r=0.056981472084815854"
	requestURL := "https://shop.48.cn/TOrder/ticket_Add"

	requestData = fmt.Sprintf(requestData, ticket.TicketID, ticket.Num, ticket.SeatType, ticket.Brand)

	// 创建一个HTTP请求客户端
	client := &http.Client{
		Timeout: time.Second,
	}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(requestData))
	if err != nil {
		log.Println(err)
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
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	var bodyMessage Message
	err = json.Unmarshal(body, &bodyMessage)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(bodyMessage)
}

type Message struct {
	HasError     bool   `json:"HasError"`
	ErrorCode    string `json:"ErrorCode"`
	Message      string `json:"Message"`
	ReturnObject string `json:"ReturnObject"`
}
