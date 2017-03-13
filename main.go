package main

import (
	"api_gateway/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	ERR_MSG         = "{\"retcode\":500, \"errmsg\": \"Internal Error!\"}"
	NOT_FOUND       = "{\"retcode\":404, \"errmsg\": \"Not Found!\"}"
	GATEWAY_EXAMPLE = `
	{
	  "proxies": [
	     {
	       "name": "Gateway-用户and订单",
	       "url": "/gateway/user_order/",
	       "method": "GET",
	       "nodes":[
	    	 {
	    	  "address": "http://localhost:8080/user/",
	    	  "retname": "user",
	    	  "timeout": 200,
	    	  "headers": ""
	    	 },
	    	 {
	    	  "address": "http://localhost:8080/order/",
	    	  "retname": "order",
	    	  "timeout": 200,
	    	  "headers": ""
	    	 }
	       ]
	     }
	   ],
	  "address": "localhost:8888"
	}
	`
)

var ROUTER_TABLE common.RouteTable

func main() {
	// 启动网关服务
	startGatewayServer()
}

func startGatewayServer() {
	fmt.Println("Init Gateway Server...")
	// 1.加载网关配置信息
	gateway := &common.Gateway{}
	err := json.Unmarshal([]byte(GATEWAY_EXAMPLE), gateway)
	if err != nil {
		fmt.Println("Init Gateway Error: ", err)
		return
	}

	// 2.构建路由表
	routetable := common.RouteTable{}
	routetable.Init()
	proxies := gateway.Proxies
	for _, proxyItem := range proxies {
		routetable.AddMapping(proxyItem.Url, proxyItem.Nodes)
	}
	ROUTER_TABLE = routetable

	// 3.接收所有请求
	http.HandleFunc("/", proxyHttpServer)
	fmt.Println("Gateway Server Listen on http://" + gateway.Address)
	http.ListenAndServe(gateway.Address, nil)
}

func proxyHttpServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	url := r.RequestURI
	fmt.Println("Gateway Receive: ", url)
	nodes := ROUTER_TABLE.SelectRouter(url)

	if nodes == nil {
		w.Write([]byte(NOT_FOUND))
		return
	}

	nodeLen := len(nodes)
	merge := nodeLen > 1
	proxyResult := &common.ProxyResult{}
	proxyResult.Init()
	wg := &sync.WaitGroup{}
	wg.Add(nodeLen)
	if merge {
		for _, nodeItem := range nodes {
			go doProxy(proxyResult, nodeItem, wg)
		}
		wg.Wait()
	} else {
		// Just one node
		doProxy(proxyResult, nodes[0], wg)
	}

	// 处理返回响应体
	if proxyResult.Body() != nil {
		apiResp := common.ApiResponse{}
		apiResp.Retcode = 200
		apiResp.Res = proxyResult.Body()

		respBytes, err := json.Marshal(apiResp)
		if err != nil {
			fmt.Println("Gateway [", url, "] Marshal Failed!!", err)
			w.Write([]byte(ERR_MSG))
		} else {
			w.Write(respBytes)
		}
	} else {
		fmt.Println("Gateway [", url, "] Failed!")
		w.Write([]byte(ERR_MSG))
	}
}

// 向后端真实服务器发送请求
func doProxy(result *common.ProxyResult, node common.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	url := node.Address
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Request [", url, "] error!", err)
		result.Failed()
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println("Request [", url, "] StatusCode not 200, but ", resp.StatusCode, "!")
		result.Failed()
		return
	}

	respBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println("Request [", url, "] and read response error!", err)
		result.Failed()
		return
	}

	serverResp := &common.ServerResponse{}
	err = json.Unmarshal(respBytes, serverResp)
	if err != nil {
		fmt.Println("Request [", url, "] and unmarshal response error!", err)
		result.Failed()
		return
	}

	result.AddResponse(node.Retname, serverResp.Res)
}
