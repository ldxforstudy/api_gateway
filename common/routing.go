package common

import "sync"

type RouteTable struct {
	mappings map[string][]Node
}

// 强制增加路由
func (r *RouteTable) AddMapping(url string, nodes []Node) {
	r.mappings[url] = nodes
}

func (r *RouteTable) Init() {
	r.mappings = make(map[string][]Node, 100)
}

// 选择url对应的路由
func (r *RouteTable) SelectRouter(url string) []Node {
	if r.mappings[url] != nil {
		nodes := r.mappings[url]
		return nodes
	}
	return nil
}

// 代理返回结果
type ProxyResult struct {
	mutex   sync.Mutex
	body    map[string]map[string]string
	success bool
}

func (res *ProxyResult) Init() {
	res.mutex = sync.Mutex{}
	res.body = make(map[string]map[string]string)
	res.success = true
}

func (res *ProxyResult) AddResponse(key string, value map[string]string) {
	res.mutex.Lock()
	if res.success {
		res.body[key] = value
	}
	res.mutex.Unlock()
}

func (res *ProxyResult) Failed() {
	res.mutex.Lock()
	res.success = false
	res.mutex.Unlock()
}

func (res *ProxyResult) Body() map[string]map[string]string {
	res.mutex.Lock()
	if res.success {
		body := res.body
		res.mutex.Unlock()
		return body
	} else {
		res.mutex.Unlock()
		return nil
	}
}
