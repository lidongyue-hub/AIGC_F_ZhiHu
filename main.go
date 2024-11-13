package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"qa/conf"
	"qa/cron"
	"qa/routes"
)

//API_SECRET_KEY = "sk-zk236783********************21ea7300";
//BASE_URL = "https://flag.smarttrot.com/v1"; #智增增的base-url
// export OPENAI_API_KEY = API_SECRET_KEY
// export OPENAI_API_BASE = BASE_URL

// 设置环境变量：export OPENAI_API_KEY=sk-zk236783de*****************************ea7300
//
//	export OPENAI_API_BASE=https://flag.smarttrot.com/v1
//
// 查看环境变量：printenv
func main() {

	conf.Init()

	cron.StartSchedule()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r := routes.NewRouter()

	if err := r.Run(":8000"); err != nil {
		panic("run server error")
	}

}

/*
func main() {
	conf.Init()
	cron.StartSchedule()

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动 HTTP 服务器
	go func() {
		defer wg.Done()

		r := routes.NewRouter()
		if err := r.Run(":8000"); err != nil {
			panic("run server error")
		}
	}()

	fmt.Println("Response Stathvjujus:")

	// 停止服务器
	// 这里可以添加一些逻辑来触发服务器的停止，例如发送一个信号给服务器
	// 然后等待服务器停止完成
	// 最后调用 wg.Done() 来告知主程序已经完成等待
	wg.Wait()
}
*/

/*
func main() {
	conf.Init()
	cron.StartSchedule()

	// 启动 HTTP 服务器
	r := routes.NewRouter()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.Run(":8000"); err != nil {
			panic("run server error")
		}
	}()

	// 发送 HTTP 请求 登陆
	url := "http://localhost:8000/api/v1/login" // 修改为实际的请求地址

	requestBody, err := json.Marshal(map[string]string{
		"username": "example_user",
		"password": "example_password",
	})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))

	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	fmt.Println("Response Statusds宋丹丹送到s:")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println("Response Statusds宋丹丹送到s的分公司公司:")
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)

	/*
		// 解析 JSON 响应
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		// 提取 token 并输出
		var jwtToken string
		if data, ok := result["data"].(map[string]interface{}); ok {
			if token, ok := data["token"].(string); ok {
				jwtToken = token
				fmt.Println("Token:", token)
			} else {
				fmt.Println("Token not found in the response")
			}
		} else {
			fmt.Println("Data not found in the response")
		}
		fmt.Println("Response jwtToken:", jwtToken)
*/
/*
		// 发送 HTTP 请求 回答问题
		url1 := "http://localhost:8000/api/v1/questions/2/answers/1" // 修改为实际的请求地址  还有问题的ID

		requestBody1, err := json.Marshal(map[string]string{ // 这里的content可以通过LLM接口来生成，
			"content": "这是我的回答内容assa",
		})
		if err != nil {
			panic(err)
		}

		req1, err := http.NewRequest("PUT", url1, bytes.NewBuffer(requestBody1))
		if err != nil {
			panic(err)
		}

		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Token", jwtToken)

		client1 := &http.Client{}
		resp1, err := client1.Do(req1)
		if err != nil {
			panic(err)
		}
		// defer resp1.Body.Close()

		fmt.Println("Response Status:", resp1.Status)


	wg.Wait()
}
*/
