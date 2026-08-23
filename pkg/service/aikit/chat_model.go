package aikit

import (
	"fmt"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/framekit"
	"github.com/example/go-ai-scaffold/pkg/library/httpkit"
	"github.com/example/go-ai-scaffold/pkg/library/jsonkit"
	"github.com/example/go-ai-scaffold/pkg/service/aikit/schema"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
)

// ChatModelConfig api连接配置信息
type ChatModelConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`

	MaxTokens int `json:"max_tokens,omitempty"`
	Timeout   int `json:"timeout"` // seconds
}

type ChatModelClient struct {
	Config ChatModelConfig
}

func NewChatModelClient(config ChatModelConfig) *ChatModelClient {
	if config.APIKey == "" {
		panic(exception.New("api key is nil"))
	}
	if config.BaseURL == "" {
		panic(exception.New("api baseUrl is nil"))
	}
	if config.Model == "" {
		panic(exception.New("model is nil"))
	}
	return &ChatModelClient{
		Config: config,
	}
}

func (client *ChatModelClient) Request(messages []schema.Message) (*schema.ResSession, *schema.ResUsage) {
	decoder := newApiResDecoder()
	req := schema.RequestBody{
		Model:         client.Config.Model,
		Messages:      messages,
		Stream:        true,
		StreamOptions: schema.StreamOption{IncludeUsage: true},
	}
	// P0 修复：overChan 改带缓冲 + notifyOver 幂等通知——
	// 流结束、请求出错、panic 兜底三条路径都可能触发通知，重复通知不能再阻塞/panic。
	overChan := make(chan bool, 1)
	notifyOver := func() {
		select {
		case overChan <- true:
		default:
		}
	}
	finalRes := &schema.ResSession{}
	var finalUsage *schema.ResUsage
	decoder.Recv(func(bytes []byte, over bool) {
		res := &schema.ResponseBody{}
		if len(bytes) > 0 {
			logkit.Debug(string(bytes))
			err := jsonkit.Unmarshal(bytes, res)
			if err != nil {
				logkit.Error(err.Error())
				return
			}
			if len(res.Choices) > 0 && res.Choices[0].Delta != nil {
				finalRes.Content += res.Choices[0].Delta.Content
				finalRes.ReasoningContent += res.Choices[0].Delta.ReasoningContent
			}
			if res.Usage != nil {
				finalUsage = res.Usage
			}
		}
		if over {
			notifyOver()
		}
	})
	go func() {
		// P0 修复：
		// ① goroutine 内 panic 无人 recover 会直接打崩整个进程（gin 的 Recover
		//    中间件管不到业务自起的 goroutine），defer recover 兜底并通知调用方；
		// ② 原用 httpkit.Request（错误即 panic）且未设 Timeout——流挂起时
		//    <-overChan 永久阻塞（goroutine + 调用方双泄漏）。改用 RequestE
		//    返回错误，并携带超时（配置缺省时兜底 300s）。
		defer func() {
			if err := recover(); err != nil {
				logkit.ErrorException(exception.New(fmt.Sprint(err), 3))
				notifyOver()
			}
		}()
		timeout := client.Config.Timeout
		if timeout <= 0 {
			timeout = 300
		}
		_, _, err := httpkit.RequestE(httpkit.Req{
			Url:         client.Config.BaseURL,
			Method:      "post",
			ContentType: "application/json",
			Header: map[string]string{
				"Authorization": "Bearer " + client.Config.APIKey,
			},
			JsonData: req,
			Stream:   true,
			Timeout:  timeout,
			StreamHandler: func(data []byte) {
				decoder.Put(data)
			},
		})
		if err != nil {
			logkit.Error("chat model request failed: " + err.Error())
			notifyOver()
		}
	}()

	<-overChan
	return finalRes, finalUsage
}

func newApiResDecoder() *framekit.Decoder {
	return framekit.NewDecoder(1024, func(bytes []byte) ([]byte, []byte, bool) {
		// 百炼的格式： data: {} ; data: [DONE]
		i := 0
		// 找到json字符串起点
		beginFlag := 0
		// 存放json首尾标记符
		jsonFlags := make([]byte, 0, 10)
		for {
			if beginFlag == 0 {
				// 寻找data:
				if len(bytes) <= i+6+6 {
					break
				}
				// 结束
				if string(bytes[i:i+6+6]) == "data: [DONE]" {
					return bytes, nil, true
				}
				if string(bytes[i:i+6]) == "data: " {
					i += 6
					beginFlag = i
					continue
				}
			} else {
				if i >= len(bytes) {
					break
				}
				switch bytes[i] {
				case '[', '{':
					// 排除作为内容含义的[{
					if (len(jsonFlags) > 0 && jsonFlags[len(jsonFlags)-1] != '"') || len(jsonFlags) == 0 {
						jsonFlags = append(jsonFlags, bytes[i])
					}
				case '"':
					if len(jsonFlags) > 0 && jsonFlags[len(jsonFlags)-1] != '"' {
						jsonFlags = append(jsonFlags, bytes[i])
					} else if len(jsonFlags) > 0 && jsonFlags[len(jsonFlags)-1] == '"' {
						jsonFlags = jsonFlags[:len(jsonFlags)-1]
					}
				case ']':
					if len(jsonFlags) > 0 && jsonFlags[len(jsonFlags)-1] == '[' {
						jsonFlags = jsonFlags[:len(jsonFlags)-1]
					}
				case '}':
					if len(jsonFlags) > 0 && jsonFlags[len(jsonFlags)-1] == '{' {
						jsonFlags = jsonFlags[:len(jsonFlags)-1]
					}
					// 识别是否结束
					if len(jsonFlags) == 0 {
						if i+1 == len(bytes) {
							return []byte{}, bytes[beginFlag : i+1], false
						}
						return bytes[i:], bytes[beginFlag : i+1], false
					}
				}
			}
			i++
		}
		return bytes, nil, false
	})
}
