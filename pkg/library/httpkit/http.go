package httpkit

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/example/go-ai-scaffold/pkg/class/const/httpconst"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/jsonkit"
)

// Doer 抽象 http.Client 的 Do 方法，便于测试 mock。
// 通过 DefaultDoer 变量注入，可在测试中替换为自定义实现。
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

var (
	// DefaultDoer 默认请求执行器。测试时可替换为 mock 实现以避免真实网络调用。
	// 注意：Req.InsecureSkipVerify=true 时绕过 DefaultDoer 直接使用 insecureClient。
	DefaultDoer Doer
	// 预构建两个 client：secureClient 校验证书（默认），insecureClient 跳过校验。
	secureClient   *http.Client
	insecureClient *http.Client
)

func init() {
	secureClient = newClient(false)
	insecureClient = newClient(true)
	// P3: 默认 Doer 指向共享 secureClient，便于测试替换
	DefaultDoer = secureClient
}

// newClient 构造配置好连接池与 cookiejar 的 http.Client。
// P1: 显式设置连接池参数，避免默认 MaxIdleConnsPerHost=2 在高并发下成为瓶颈。
func newClient(skipTLS bool) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: skipTLS},
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &http.Client{Transport: tr, Jar: jar}
}

// Req 单次请求参数。
//
// body 字段互斥（M4）：BinaryData / JsonData / FormData 同时只能设置一个，违者 panic。
// QueryData 与 body 正交（M1）：可独立设置，不参与 body 构造。
// 下载场景：Stream=true 配合 StreamHandler 流式回调；OutputFilePath 落盘；
// 二者同时设置时 file 优先，StreamHandler 不被调用（M2：行为明确为 file 优先）。
type Req struct {
	Method      string
	Url         string
	Header      map[string]string
	ContentType string          // 留空时按 body 类型自动推断；用户已设 Header["Content-Type"] 时尊重用户值
	Ctx         context.Context // A1：每请求独立 context，用于超时与取消。nil 时使用 context.Background

	// body（互斥）
	BinaryData []byte            // 原始字节
	JsonData   any               // JSON
	FormData   map[string]string // 表单

	// query（与 body 正交）
	QueryData map[string]string

	// 下载
	OutputFilePath string            // 落盘路径；为空时返回响应体字节
	Append         bool              // OutputFilePath 已存在时是否追加，默认 truncate（M3）
	Stream         bool              // 流式读取
	StreamHandler  func(data []byte) // 流式回调（与 OutputFilePath 互斥，file 优先）

	// 控制
	Timeout            int  // 超时秒数，>0 时通过 context.WithTimeout 覆盖 Ctx
	InsecureSkipVerify bool // 单请求级别跳过 TLS 校验；true 时使用 insecureClient
}

// Request 阻塞发起请求，错误时 panic。
// A2: 返回 []byte 替代 string，避免二进制响应的 UTF-8 编码问题。
// 流式 / 落盘场景返回 (nil, statusCode)。
func Request(reqParams Req) ([]byte, int) {
	body, c, err := RequestE(reqParams)
	if err != nil {
		panic(exception.New(err.Error()))
	}
	return body, c
}

// RequestE 阻塞发起请求，返回响应字节、状态码与错误。
func RequestE(reqParams Req) ([]byte, int, error) {
	req, err := genRequest(reqParams)
	if err != nil {
		return nil, 0, err
	}

	// H1: 超时通过 context 传递，避免并发修改共享 client.Timeout 造成竞态
	ctx := reqParams.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if reqParams.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(reqParams.Timeout)*time.Second)
		defer cancel()
	}
	req = req.WithContext(ctx)

	// H2: InsecureSkipVerify 单请求级别跳过 TLS 校验，绕过 DefaultDoer 直接用 insecureClient
	doer := DefaultDoer
	if reqParams.InsecureSkipVerify {
		doer = insecureClient
	}
	resp, err := doer.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if reqParams.Stream {
		return readStream(resp.Body, reqParams, resp.StatusCode)
	}
	return readAll(resp.Body, reqParams, resp.StatusCode)
}

// RequestJson A3: 发起请求并将 JSON 响应反序列化到 resp，省去调用方手动 Unmarshal。
func RequestJson(reqParams Req, resp any) error {
	body, _, err := RequestE(reqParams)
	if err != nil {
		return err
	}
	return jsonkit.Unmarshal(body, resp)
}

// genRequest 构造 http.Request。
// H3: 不 panic，错误正常返回，由调用方统一处理。
func genRequest(reqParams Req) (*http.Request, error) {
	if reqParams.Method == "" {
		reqParams.Method = http.MethodPost
	}
	reqParams.Method = strings.ToUpper(reqParams.Method)

	bodyReader, defaultContentType, err := buildBody(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(reqParams.Method, reqParams.Url, bodyReader)
	if err != nil {
		return nil, err
	}

	// M1: QueryData 与 body 正交，独立拼接
	if len(reqParams.QueryData) > 0 {
		q := req.URL.Query()
		for k, v := range reqParams.QueryData {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Q1: 先应用用户 Header，缺省时再补默认 Content-Type，避免覆盖用户显式设置
	for k, v := range reqParams.Header {
		req.Header.Set(k, v)
	}
	if reqParams.ContentType != "" {
		req.Header.Set("Content-Type", reqParams.ContentType)
	} else if defaultContentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", defaultContentType)
	}

	return req, nil
}

// buildBody 根据 Req 选择 body 类型，返回 reader 与默认 Content-Type。
// M4: BinaryData/JsonData/FormData 互斥，同设多个时 panic。
// P2: JsonData 直接 Marshal 为 []byte，避免 string→[]byte 重复拷贝。
func buildBody(reqParams Req) (io.Reader, string, error) {
	setCount := 0
	if reqParams.BinaryData != nil {
		setCount++
	}
	if reqParams.JsonData != nil {
		setCount++
	}
	if reqParams.FormData != nil {
		setCount++
	}
	if setCount > 1 {
		return nil, "", exception.New("BinaryData/JsonData/FormData 互斥，同时只能设置一个")
	}

	if reqParams.BinaryData != nil {
		return bytes.NewBuffer(reqParams.BinaryData), "", nil
	}
	if reqParams.JsonData != nil {
		data, err := jsonkit.Marshal(reqParams.JsonData)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewBuffer(data), httpconst.ContentTypeJSON, nil
	}
	if reqParams.FormData != nil {
		form := make(url.Values)
		for k, v := range reqParams.FormData {
			form.Add(k, v)
		}
		return strings.NewReader(form.Encode()), httpconst.ContentTypeForm, nil
	}
	return nil, "", nil
}

// readStream 流式读取响应体，按 OutputFilePath/StreamHandler 处理。
func readStream(body io.Reader, reqParams Req, statusCode int) ([]byte, int, error) {
	reader := bufio.NewReader(body)
	buffer := make([]byte, 4*1024)
	var fout *os.File
	if reqParams.OutputFilePath != "" {
		var err error
		fout, err = openFile(reqParams.OutputFilePath, reqParams.Append)
		if err != nil {
			return nil, 0, err
		}
		defer fout.Close()
	}
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, 0, err
		}
		if n > 0 {
			if fout != nil {
				if _, werr := fout.Write(buffer[:n]); werr != nil {
					return nil, 0, werr
				}
			} else if reqParams.StreamHandler != nil {
				// M2: file 优先，未设 OutputFilePath 时才回调
				reqParams.StreamHandler(buffer[:n])
			}
		}
		if err == io.EOF {
			break
		}
		if n == 0 {
			// Q2: 兜底防忙等，bufio.Reader.Read 通常会阻塞，n==0 实际少见
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil, statusCode, nil
}

// readAll 一次性读取响应体，按 OutputFilePath 决定落盘或返回字节。
func readAll(body io.Reader, reqParams Req, statusCode int) ([]byte, int, error) {
	if reqParams.OutputFilePath != "" {
		fout, err := openFile(reqParams.OutputFilePath, reqParams.Append)
		if err != nil {
			return nil, 0, err
		}
		defer fout.Close()
		if _, err := io.Copy(fout, body); err != nil {
			return nil, 0, err
		}
		return nil, statusCode, nil
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, 0, err
	}
	return data, statusCode, nil
}

// openFile M3: 移除 TOCTOU 的 Exists 判断，默认 truncate；Append=true 时追加。
// 修复前：Exists 为 true 时追加，重试下载会得到损坏文件（旧前半 + 新后半）。
func openFile(fp string, append bool) (*os.File, error) {
	flag := os.O_WRONLY | os.O_CREATE
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	return os.OpenFile(fp, flag, 0644)
}
