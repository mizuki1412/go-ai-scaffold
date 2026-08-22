package configkey

const JwtSecretKey = "jwt.secret"
const JwtExpire = "jwt.expire"

// JwtIdle 会话空闲窗口/小时：AuthJWT 每次鉴权通过即重置白名单 TTL（滑动续期），
// 空闲超时未请求即登录失效；<=0 时退化为不滑动（窗口=JwtExpire）。
const JwtIdle = "jwt.idle"
