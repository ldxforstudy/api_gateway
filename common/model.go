package common

type Gateway struct {
	Address string  `json:"address"`
	Proxies []Proxy `json:"proxies"`
}

type Proxy struct {
	Name   string `json:"name"`
	Url    string `json:"url"`
	Method string `json:"method"`
	Nodes  []Node `json:"nodes"`
}

type Node struct {
	Address string `json:"address"`
	Retname string `json:"retname"`
	Timeout int    `json:"timeout"`
	Headers string `json:"headers"`
}

type ApiResponse struct {
	Retcode int `json:"retcode"`
	Res map[string]map[string]string `json:"res"`
}

type ServerResponse struct {
	Retcode int `json:"retcode"`
	Res map[string]string `json:"res"`
}