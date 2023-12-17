package rest

import (
	"github.com/rulego/rulego"
	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/components/action"
	"github.com/rulego/rulego/endpoint"
	"github.com/rulego/rulego/test"
	"github.com/rulego/rulego/test/assert"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var testdataFolder = "../../testdata"
var testServer = ":9090"
var testConfigServer = ":9091"

// 测试请求/响应消息
func TestRestMessage(t *testing.T) {
	t.Run("Request", func(t *testing.T) {
		var request = &RequestMessage{}
		test.EndpointMessage(t, request)
	})
	t.Run("Response", func(t *testing.T) {
		var response = &ResponseMessage{}
		test.EndpointMessage(t, response)
	})
}

func TestRestEndpointConfig(t *testing.T) {
	config := rulego.NewConfig(types.WithDefaultPool())
	//创建rest endpoint服务
	epStarted, _ := endpoint.New(Type, config, Config{
		Server: testConfigServer,
	})
	assert.Equal(t, testConfigServer, epStarted.Id())

	go func() {
		err := epStarted.Start()
		assert.Equal(t, "http: Server closed", err.Error())
	}()

	time.Sleep(time.Millisecond * 200)

	epErr, _ := endpoint.New(Type, config, types.Configuration{
		"server": testConfigServer,
	})
	_, err := epErr.AddRouter(nil, "POST")
	assert.Equal(t, "router can not nil", err.Error())
	//err := epErr.Start()
	//assert.NotNil(t, err)
	ep, _ := endpoint.New(Type, config, types.Configuration{
		"server": testConfigServer,
	})

	assert.Equal(t, testConfigServer, ep.Id())
	//_, err := ep.AddRouter(nil)
	//assert.Equal(t, "router can not nil", err.Error())
	testUrl := "/api/test"
	router := endpoint.NewRouter().From(testUrl).End()
	_, err = ep.AddRouter(router)
	assert.Equal(t, "need to specify HTTP method", err.Error())

	router = endpoint.NewRouter().From(testUrl).End()
	routerId, err := ep.AddRouter(router, "POST")
	assert.Equal(t, "/api/test", routerId)

	restEndpoint, ok := ep.(*Rest)
	assert.True(t, ok)

	router = endpoint.NewRouter().From(testUrl).End()
	restEndpoint.POST(router)
	restEndpoint.GET(router)
	restEndpoint.DELETE(router)
	restEndpoint.PATCH(router)
	restEndpoint.OPTIONS(router)
	restEndpoint.HEAD(router)
	restEndpoint.PUT(router)

	//删除路由
	ep.RemoveRouter(routerId)
	ep.RemoveRouter(routerId, "POST")

	epStarted.Destroy()
	epErr.Destroy()
	time.Sleep(time.Millisecond * 200)
}

func TestRestEndpoint(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	stop := make(chan struct{})
	//启动服务
	go startServer(t, stop, &wg)
	//等待服务器启动完毕
	time.Sleep(time.Millisecond * 200)

	config := rulego.NewConfig(types.WithDefaultPool())
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err2 error) {
		assert.Equal(t, "ok", msg.Data)
	})
	metaData := types.BuildMetadata(make(map[string]string))
	msg1 := ctx.NewMsg("TEST_MSG_TYPE_AA", metaData, "{\"name\":\"lala\"}")

	sendMsg(t, "http://127.0.0.1"+testServer+"/api/v1/msg2Chain2/TEST_MSG_TYPE1?aa=xx", "POST", msg1, ctx)

	//停止服务器
	stop <- struct{}{}
	time.Sleep(time.Millisecond * 200)
	wg.Wait()
}

// 发送消息到rest服务器
func sendMsg(t *testing.T, url, method string, msg types.RuleMsg, ctx types.RuleContext) types.Node {
	node, _ := rulego.Registry.NewNode("restApiCall")
	var configuration = make(types.Configuration)
	configuration["restEndpointUrlPattern"] = url
	configuration["requestMethod"] = method
	config := types.NewConfig()
	err := node.Init(config, configuration)
	if err != nil {
		t.Fatal(err)
	}
	//发送消息
	node.OnMsg(ctx, msg)
	return node
}

// 启动rest服务
func startServer(t *testing.T, stop chan struct{}, wg *sync.WaitGroup) {
	buf, err := os.ReadFile(testdataFolder + "/chain_msg_type_switch.json")
	if err != nil {
		t.Fatal(err)
	}
	config := rulego.NewConfig(types.WithDefaultPool())
	//注册规则链
	_, _ = rulego.New("default", buf, rulego.WithConfig(config))

	//创建http endpoint服务
	restEndpoint, err := endpoint.New(Type, config, Config{
		Server: testServer,
	})

	go func() {
		for {
			select {
			case <-stop:
				// 接收到中断信号，退出循环
				restEndpoint.Destroy()
				return
			default:
			}
		}
	}()
	//添加全局拦截器
	restEndpoint.AddInterceptors(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		//权限校验逻辑
		return true
	})
	//设置跨域
	restEndpoint.(*Endpoint).GlobalOPTIONS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			// 设置 CORS 相关的响应头
			header := w.Header()
			header.Set("Access-Control-Allow-Methods", r.Header.Get("Allow"))
			header.Set("Access-Control-Allow-Headers", "*")
			header.Set("Access-Control-Allow-Origin", "*")
		}
		// 返回 204 状态码
		w.WriteHeader(http.StatusNoContent)
	}))
	//路由1
	router1 := endpoint.NewRouter().From("/api/v1/hello/:name").Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		//处理请求
		request, ok := exchange.In.(*RequestMessage)
		if ok {
			if request.request.Method != http.MethodGet {
				//响应错误
				exchange.Out.SetStatusCode(http.StatusMethodNotAllowed)
				//不执行后续动作
				return false
			} else {
				//响应请求
				exchange.Out.Headers().Set(ContentTypeKey, JsonContextType)
				exchange.Out.SetBody([]byte(exchange.In.From() + "\n"))
				exchange.Out.SetBody([]byte("s1 process" + "\n"))
				name := request.GetMsg().Metadata.GetValue("name")
				if name == "break" {
					//不执行后续动作
					return false
				} else {
					return true
				}

			}
		} else {
			exchange.Out.Headers().Set(ContentTypeKey, JsonContextType)
			exchange.Out.SetBody([]byte(exchange.In.From()))
			exchange.Out.SetBody([]byte("s1 process" + "\n"))
			return true
		}

	}).Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		exchange.Out.SetBody([]byte("s2 process" + "\n"))
		return true
	}).End()

	//路由2 采用配置方式调用规则链
	router2 := endpoint.NewRouter().From("/api/v1/msg2Chain1/:msgType").To("chain:default").End()

	//路由3 采用配置方式调用规则链,to路径带变量
	router3 := endpoint.NewRouter().From("/api/v1/msg2Chain2/:msgType").Transform(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		msg := exchange.In.GetMsg()
		//获取消息类型
		msg.Type = msg.Metadata.GetValue("msgType")

		//从header获取用户ID
		userId := exchange.In.Headers().Get("userId")
		if userId == "" {
			userId = "default"
		}
		//把userId存放在msg元数据
		msg.Metadata.PutValue("userId", userId)
		return true
	}).Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		requestMessage, ok := exchange.In.(*RequestMessage)
		assert.True(t, ok)
		assert.NotNil(t, requestMessage.Request())
		assert.Equal(t, JsonContextType, requestMessage.Headers().Get(ContentTypeKey))

		from := requestMessage.From()
		msgType := requestMessage.GetMsg().Metadata.GetValue("msgType")
		assert.Equal(t, "/api/v1/msg2Chain2/"+msgType+"?aa=xx", from)
		assert.Equal(t, "xx", requestMessage.GetParam("aa"))

		responseMessage, ok := exchange.Out.(*ResponseMessage)
		assert.NotNil(t, responseMessage.Response())

		assert.Equal(t, "/api/v1/msg2Chain2/"+msgType+"?aa=xx", responseMessage.From())
		assert.Equal(t, "xx", responseMessage.GetParam("aa"))
		//响应给客户端
		exchange.Out.Headers().Set(ContentTypeKey, JsonContextType)
		exchange.Out.SetStatusCode(200)
		exchange.Out.SetBody([]byte("ok"))
		return true
	}).To("chain:${userId}").Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		outMsg := exchange.Out.GetMsg()
		if outMsg != nil {
			assert.Equal(t, true, len(outMsg.Metadata.Values()) > 1)
		}
		return true
	}).End()

	//路由4 直接调用node组件方式
	router4 := endpoint.NewRouter().From("/api/v1/msgToComponent1/:msgType").Transform(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		msg := exchange.In.GetMsg()
		//获取消息类型
		msg.Type = msg.Metadata.GetValue("msgType")
		return true
	}).Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		//响应给客户端
		exchange.Out.Headers().Set(ContentTypeKey, JsonContextType)
		exchange.Out.SetBody([]byte("ok"))
		return true
	}).ToComponent(func() types.Node {
		//定义日志组件，处理数据
		var configuration = make(types.Configuration)
		configuration["jsScript"] = `
		return 'log::Incoming message:\n' + JSON.stringify(msg) + '\nIncoming metadata:\n' + JSON.stringify(metadata);
       `
		logNode := &action.LogNode{}
		_ = logNode.Init(config, configuration)
		return logNode
	}()).End()

	//路由5 采用配置方式调用node组件
	router5 := endpoint.NewRouter().From("/api/v1/msgToComponent2/:msgType").Transform(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		msg := exchange.In.GetMsg()
		//获取消息类型
		msg.Type = msg.Metadata.GetValue("msgType")
		return true
	}).Process(func(router *endpoint.Router, exchange *endpoint.Exchange) bool {
		//响应给客户端
		exchange.Out.Headers().Set(ContentTypeKey, JsonContextType)
		exchange.Out.SetBody([]byte("ok"))
		return true
	}).To("component:log", types.Configuration{"jsScript": `
		return 'log::Incoming message:\n' + JSON.stringify(msg) + '\nIncoming metadata:\n' + JSON.stringify(metadata);
       `}).End()

	//注册路由,Get 方法
	_, _ = restEndpoint.AddRouter(router1, "GET")
	//注册路由，POST方式
	_, _ = restEndpoint.AddRouter(router2, "POST")
	_, _ = restEndpoint.AddRouter(router3, "POST")
	_, _ = restEndpoint.AddRouter(router4, "POST")
	_, _ = restEndpoint.AddRouter(router5, "POST")

	assert.NotNil(t, restEndpoint.(*Endpoint).Router())
	//启动服务
	_ = restEndpoint.Start()
	wg.Done()
}
