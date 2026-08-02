package openapi

// ==================== Top-level ====================

// ApiDocV3 OpenAPI 3.1 文档根。
type ApiDocV3 struct {
	Openapi    string                                       `json:"openapi"`
	Info       *ApiDocV3Info                                `json:"info,omitempty"`
	Tags       []*ApiDocV3Tag                               `json:"tags,omitempty"`
	Paths      map[string]map[string]*ApiDocV3PathOperation `json:"paths"` // path:method:operation
	Components *ApiDocV3ComponentObj                        `json:"components,omitempty"`
	Servers    []*ApiDocV3Server                            `json:"servers,omitempty"` // 至少一个；多环境时多项
	// Security 顶层默认鉴权要求，作用于所有未显式声明 security 的 operation。
	Security []map[string][]string `json:"security,omitempty"`
	// ExternalDocs 外部文档链接（可选）
	ExternalDocs *ApiDocV3ExternalDocs `json:"externalDocs,omitempty"`
	// Webhooks 3.1 webhook 定义
	Webhooks map[string]*ApiDocV3PathItem `json:"webhooks,omitempty"`
}

// ApiDocV3PathItem 用于 paths 项值（也用于 webhooks）。当前仅承载一个 method operation。
type ApiDocV3PathItem struct {
	Get     *ApiDocV3PathOperation `json:"get,omitempty"`
	Post    *ApiDocV3PathOperation `json:"post,omitempty"`
	Put     *ApiDocV3PathOperation `json:"put,omitempty"`
	Delete  *ApiDocV3PathOperation `json:"delete,omitempty"`
	Options *ApiDocV3PathOperation `json:"options,omitempty"`
	Head    *ApiDocV3PathOperation `json:"head,omitempty"`
	Patch   *ApiDocV3PathOperation `json:"patch,omitempty"`
	Trace   *ApiDocV3PathOperation `json:"trace,omitempty"`
}

// ApiDocV3Server 服务端点。
type ApiDocV3Server struct {
	Url         string                `json:"url"`
	Description string                `json:"description,omitempty"`
	Variables   map[string]*ServerVar `json:"variables,omitempty"`
}

// ServerVar 服务端点变量（如 {protocol}、{host}）。
type ServerVar struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// ApiDocV3ExternalDocs 外部文档链接。
type ApiDocV3ExternalDocs struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url"`
}

// ==================== Info ====================

type ApiDocV3Info struct {
	Title          string               `json:"title"`
	Description    string               `json:"description,omitempty"`
	TermsOfService string               `json:"termsOfService,omitempty"`
	Contact        *ApiDocV3InfoContact `json:"contact,omitempty"`
	License        *ApiDocV3InfoLicense `json:"license,omitempty"`
	Version        string               `json:"version"`
}

type ApiDocV3InfoContact struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type ApiDocV3InfoLicense struct {
	Name       string `json:"name"`
	Url        string `json:"url,omitempty"`
	Identifier string `json:"identifier,omitempty"` // 3.1：SPDX license identifier
}

// ==================== Tag ====================

type ApiDocV3Tag struct {
	Name         string                `json:"name"`
	Description  string                `json:"description,omitempty"`
	ExternalDocs *ApiDocV3ExternalDocs `json:"externalDocs,omitempty"`
}

// ==================== Operation ====================

// ApiDocV3PathOperation 单个接口的 operation。
type ApiDocV3PathOperation struct {
	Tags        []string                    `json:"tags,omitempty"`
	Summary     string                      `json:"summary,omitempty"`
	Description string                      `json:"description,omitempty"`
	OperationId string                      `json:"operationId,omitempty"`
	Deprecated  bool                        `json:"deprecated,omitempty"`
	Parameters  []*ApiDocV3ReqParam         `json:"parameters,omitempty"`
	RequestBody *ApiDocV3ReqBody            `json:"requestBody,omitempty"`
	Responses   map[string]*ApiDocV3ResBody `json:"responses"`
	// Security operation 级安全要求；为空切片表示不鉴权（覆盖顶层）。
	Security []map[string][]string `json:"security,omitempty"`
	Servers  []*ApiDocV3Server     `json:"servers,omitempty"`
}

// ==================== Parameter ====================

// ApiDocV3ReqParam 路径/查询/头/cookie 参数。
type ApiDocV3ReqParam struct {
	Ref             string          `json:"$ref,omitempty"` // 引用 components.parameters
	Name            string          `json:"name,omitempty"`
	In              string          `json:"in,omitempty"` // query | header | path | cookie
	Description     string          `json:"description,omitempty"`
	Required        bool            `json:"required,omitempty"`
	Deprecated      bool            `json:"deprecated,omitempty"`
	AllowEmptyValue bool            `json:"allowEmptyValue,omitempty"` // 仅 query
	Style           string          `json:"style,omitempty"`           // form | simple | label | matrix | spaceDelimited | pipeDelimited | deepObject
	Explode         *bool           `json:"explode,omitempty"`         // 指针：false 与省略语义不同
	AllowReserved   bool            `json:"allowReserved,omitempty"`
	Schema          *ApiDocV3Schema `json:"schema,omitempty"`
	Example         any             `json:"example,omitempty"`
}

// ==================== Request/Response Body ====================

type ApiDocV3ReqBody struct {
	Description string                            `json:"description,omitempty"`
	Required    bool                              `json:"required,omitempty"`
	Content     map[string]*ApiDocV3SchemaWrapper `json:"content"`
}

type ApiDocV3ResBody struct {
	Ref         string                            `json:"$ref,omitempty"` // 引用 components.responses
	Description string                            `json:"description,omitempty"`
	Content     map[string]*ApiDocV3SchemaWrapper `json:"content,omitempty"`
	Headers     map[string]*ApiDocV3Header        `json:"headers,omitempty"`
}

type ApiDocV3Header struct {
	Description string          `json:"description,omitempty"`
	Required    bool            `json:"required,omitempty"`
	Schema      *ApiDocV3Schema `json:"schema,omitempty"`
}

type ApiDocV3SchemaWrapper struct {
	Schema  *ApiDocV3Schema `json:"schema,omitempty"`
	Example any             `json:"example,omitempty"`
}

// ==================== Schema ====================

// ApiDocV3Schema JSON Schema 子集。
// Type 为切片以支持 3.1 的 type:["string","null"]；序列化时若只含一个元素，输出标量以兼容旧 UI。
type ApiDocV3Schema struct {
	Ref                  string                     `json:"$ref,omitempty"`
	Type                 any                        `json:"type,omitempty"` // string 或 []string
	Format               string                     `json:"format,omitempty"`
	Title                string                     `json:"title,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Default              any                        `json:"default,omitempty"`
	Example              any                        `json:"example,omitempty"`
	Enum                 []any                      `json:"enum,omitempty"`
	Properties           map[string]*ApiDocV3Schema `json:"properties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	Items                *ApiDocV3Schema            `json:"items,omitempty"`
	AdditionalProperties *ApiDocV3Schema            `json:"additionalProperties,omitempty"` // 布尔或 schema；此处用 schema，false 用 ExclusiveMaximum=false 表达
	Nullable             bool                       `json:"nullable,omitempty"`             // 3.0 兼容；3.1 通过 Type 数组表达
	ReadOnly             bool                       `json:"readOnly,omitempty"`
	WriteOnly            bool                       `json:"writeOnly,omitempty"`
	Deprecated           bool                       `json:"deprecated,omitempty"`
	Minimum              *float64                   `json:"minimum,omitempty"`
	Maximum              *float64                   `json:"maximum,omitempty"`
	MinLength            *int64                     `json:"minLength,omitempty"`
	MaxLength            *int64                     `json:"maxLength,omitempty"`
	Pattern              string                     `json:"pattern,omitempty"`
	MinItems             *int64                     `json:"minItems,omitempty"`
	MaxItems             *int64                     `json:"maxItems,omitempty"`
	UniqueItems          bool                       `json:"uniqueItems,omitempty"`
	MinProperties        *int64                     `json:"minProperties,omitempty"`
	MaxProperties        *int64                     `json:"maxProperties,omitempty"`
	ContentMediaType     string                     `json:"contentMediaType,omitempty"`
	OneOf                []*ApiDocV3Schema          `json:"oneOf,omitempty"`
	AnyOf                []*ApiDocV3Schema          `json:"anyOf,omitempty"`
	AllOf                []*ApiDocV3Schema          `json:"allOf,omitempty"`
	Not                  *ApiDocV3Schema            `json:"not,omitempty"`
}

// ==================== Components ====================

type ApiDocV3ComponentObj struct {
	Schemas         map[string]*ApiDocV3Schema         `json:"schemas,omitempty"`
	Parameters      map[string]*ApiDocV3ReqParam       `json:"parameters,omitempty"`
	Headers         map[string]*ApiDocV3Header         `json:"headers,omitempty"`
	RequestBodies   map[string]*ApiDocV3ReqBody        `json:"requestBodies,omitempty"`
	Responses       map[string]*ApiDocV3ResBody        `json:"responses,omitempty"`
	SecuritySchemes map[string]*ApiDocV3SecurityScheme `json:"securitySchemes,omitempty"`
	Examples        map[string]*ApiDocV3Example        `json:"examples,omitempty"`
}

// ApiDocV3Example 复用的示例对象。
type ApiDocV3Example struct {
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Value       any    `json:"value,omitempty"`
}

// ApiDocV3SecurityScheme 安全方案。当前仅实现 JWT Bearer。
type ApiDocV3SecurityScheme struct {
	Type         string `json:"type"`                   // http | apiKey | oauth2 | openIdConnect
	Scheme       string `json:"scheme,omitempty"`       // bearer | basic | digest
	BearerFormat string `json:"bearerFormat,omitempty"` // JWT
	Description  string `json:"description,omitempty"`
	// apiKey 专用
	In   string `json:"in,omitempty"`   // header | query | cookie
	Name string `json:"name,omitempty"` // 字段名
}

// ==================== 常量 ====================

const (
	SchemaTypeInteger string = "integer"
	SchemaTypeString         = "string"
	SchemaTypeNumber         = "number"
	SchemaTypeBool           = "boolean"
	SchemaTypeObject         = "object"
	SchemaTypeArray          = "array"

	SchemaFormatInt32    = "int32"
	SchemaFormatInt64    = "int64"
	SchemaFormatFloat    = "float"
	SchemaFormatDouble   = "double"
	SchemaFormatByte     = "byte"
	SchemaFormatBinary   = "binary"
	SchemaFormatDate     = "date"
	SchemaFormatDateTime = "date-time"

	ParamInQuery  = "query"
	ParamInHeader = "header"
	ParamInPath   = "path"
	ParamInCookie = "cookie"

	// SecuritySchemeBearer JWT Bearer scheme 名
	SecuritySchemeBearer = "bearer"
)
