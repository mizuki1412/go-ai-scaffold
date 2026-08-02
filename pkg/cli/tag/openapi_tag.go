package tag

// ParamIn openapi的参数路径in参数
var ParamIn = Tag{Name: "in"}

// Schema openapi schema配置
var Schema = Tag{Name: "schema"}

// SchemaIgnore openapi不序列化
const SchemaIgnore = "ignore"

// RetData 专用于RestRet，自定义时表示data所在的区域
var RetData = Tag{Name: "data", Value: "true"}

// Example 字段/参数示例值
var Example = Tag{Name: "example"}

// Deprecated 标记字段/参数/operation 为已弃用
var Deprecated = Tag{Name: "deprecated", Value: "true"}

// ReadOnly 标记字段为只读（响应中出现，请求中忽略）
var ReadOnly = Tag{Name: "readOnly", Value: "true"}

// WriteOnly 标记字段为只写（请求中出现，响应中忽略）
var WriteOnly = Tag{Name: "writeOnly", Value: "true"}

// AllowEmptyValue 允许空值（仅 query 参数）
var AllowEmptyValue = Tag{Name: "allowEmptyValue", Value: "true"}

// Nullable 标记字段可为 null
var Nullable = Tag{Name: "nullable", Value: "true"}
