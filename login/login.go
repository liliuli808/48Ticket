package login

import (
	"awesomeProject/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func QuickLogin(ticket model.TicketType) string {
	// 目标登录URL
	url := "https://user.48.cn/QuickLogin/login/"

	if ticket.Cookie != "" {
		return ticket.Cookie
	}

	// 构造登录表单数据

	payload := strings.NewReader(fmt.Sprintf("phone=&phonecode=&login_type=&area=&preg=&username=%s&password=%s", ticket.Username, ticket.Password))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Origin", "https://user.48.cn")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://user.48.cn/Login/index.html?return_url=https://shop.48.cn/home/index")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", `"Linux"`)

	client := &http.Client{}
	resp, err := client.Do(req)

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
	}

	srcContent := extractSrcContent(response.Desc)
	cookiesStr := ""
	for _, v := range srcContent {
		req, err := http.NewRequest("GET", v, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return ""
		}
		req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Add("Origin", "https://user.48.cn")
		req.Header.Add("Pragma", "no-cache")
		req.Header.Add("Referer", "https://user.48.cn/Login/index.html?return_url=https://shop.48.cn/home/index")
		req.Header.Add("Sec-Fetch-Dest", "empty")
		req.Header.Add("Sec-Fetch-Mode", "cors")
		req.Header.Add("Sec-Fetch-Site", "same-origin")
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.Header.Add("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
		req.Header.Add("sec-ch-ua-mobile", "?0")
		req.Header.Add("sec-ch-ua-platform", `"Linux"`)

		client := &http.Client{}
		resp, err := client.Do(req)

		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name != ".AspNet.ApplicationCookie" {
				continue
			}
			cookiesStr += fmt.Sprintf("  %s=%s;", cookie.Name, cookie.Value, cookie.Domain)
		}
	}
	return cookiesStr
}

func extractSrcContent(input string) []string {
	// Regular expression pattern to match script tags and capture src attribute
	scriptPattern := `<script[^>]*\s+src="([^"]+)"[^>]*>`

	// Compile the regular expression
	re := regexp.MustCompile(scriptPattern)

	// Find all matches in the input
	matches := re.FindAllStringSubmatch(input, -1)

	// Extract content from the captured group
	var srcContents []string
	for _, match := range matches {
		if len(match) == 2 {
			srcContents = append(srcContents, match[1])
		}
	}

	return srcContents
}

type Response struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
}
