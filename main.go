package main

import (
	"awesomeProject/login"
	"awesomeProject/model"
	"bytes"
	"encoding/json"
	"fmt"
	_ "go.uber.org/automaxprocs"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	// 创建或打开一个日志文件
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer logFile.Close()

	// 设置日志输出到文件
	log.SetOutput(logFile)

	// 读取YAML配置文件
	yamlFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("无法读取YAML文件: %v", err)
	}

	// 创建TicketType结构体实例
	var ticket model.TicketType

	// 解析YAML配置文件到结构体
	err = yaml.Unmarshal(yamlFile, &ticket)
	if err != nil {
		log.Fatalf("无法解析YAML文件: %v", err)
	}
	ticket.Cookie = login.QuickLogin(ticket)
	//// 设置最大并发请求数
	maxConcurrentRequests := 3

	// 创建通道用于通知抢票结果
	for {
		format := "2006-01-02 15:04:05"

		var cstZone = time.FixedZone("CST", 8*3600) // 东八
		time.Local = cstZone

		// 使用 time 包中的 Parse 函数将字符串解析为时间对象
		dateTime, err := time.ParseInLocation(format, ticket.StartTime, cstZone)
		if err != nil {
			fmt.Println("解析日期时间失败:", err)
			continue
		}
		currentTime := time.Now().Add(20 * time.Millisecond)
		// 如果当前时间早于抢票时间，则等待
		if currentTime.Before(dateTime) {
			continue
		}
		var wg sync.WaitGroup
		// 启动最大并发请求数的抢票任务
		for i := 0; i < maxConcurrentRequests; i++ {
			wg.Add(1)
			go func() {
				ticketAdd(ticket)
				defer wg.Done()
			}()
		}
		wg.Wait()
		time.Sleep(time.Second)
	}

}

func getIndexTickets() TicketList {
	url := "https://shop.48.cn/Home/IndexTickets"
	var ticketList TicketList

	payload := []byte("brand_id=-1&team_type=-1&date_type=0")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ticketList
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", "route=46bf5666bf631c4b75db945ee8ab5a17; Hm_lvt_f32737cfa62ed971bb3185792d3204eb=1705209344,1705395989,1705403887,1706001331; Hm_lpvt_f32737cfa62ed971bb3185792d3204eb=1706001331")
	req.Header.Set("Origin", "https://shop.48.cn")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://shop.48.cn/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return ticketList
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ticketList
	}
	json.Unmarshal(body, &ticketList)
	return ticketList
}

type TicketList struct {
	Listnew []struct {
		TicketsId       int    `json:"tickets_id"`
		TheatreId       int    `json:"theatre_id"`
		BrandId         int    `json:"brand_id"`
		IsOnSale        bool   `json:"is_on_sale"`
		StartPlaydate   string `json:"start_playdate"`
		TicketsName     string `json:"tickets_name"`
		TeamType        int    `json:"team_type"`
		TeamTypeBrandId int    `json:"team_type_brand_id"`
		MerchandiseType int    `json:"merchandise_type"`
		TicketsSales    []struct {
			SeatType     int     `json:"seat_type"`
			Amount       int     `json:"amount"`
			Price        float64 `json:"price"`
			RankIntegral int     `json:"rank_integral"`
		} `json:"tickets_sales"`
	} `json:"listnew"`
	FlashList []interface{} `json:"flash_list"`
}

func ticketAdd(ticket model.TicketType) bool {
	log.Println("开始抢票")
	requestData := "ticketsid=%s&num=%s&seattype=%s&brand_id=%s&choose_times_end=-1&ticketstype=2&r=0.056981472084815854"
	requestURL := "https://shop.48.cn/TOrder/ticket_Add"

	requestData = fmt.Sprintf(requestData, ticket.TicketID, ticket.Num, ticket.SeatType, ticket.Brand)

	// 创建一个HTTP请求客户端
	client := &http.Client{}

	// 创建一个HTTP POST请求
	request, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(requestData))
	if err != nil {
		log.Println(err)
		return false
	}

	// 设置请求头
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Cookie", ticket.Cookie)
	request.Header.Set("Origin", "https://shop.48.cn")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Referer", fmt.Sprintf("https://shop.48.cn/tickets/item/%s?seat_type=%s", ticket.TicketID, ticket.SeatType))
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("sec-ch-ua-platform", `"Linux"`)

	// 发送HTTP请求
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	var bodyMessage Message
	err = json.Unmarshal(body, &bodyMessage)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println(string(body))
	if bodyMessage.HasError == true {
		return false
	}
	if bodyMessage.Message == "success" {
		return true
	}
	return false
}

type Message struct {
	HasError     bool   `json:"HasError"`
	ErrorCode    string `json:"ErrorCode"`
	Message      string `json:"Message"`
	ReturnObject string `json:"ReturnObject"`
}
