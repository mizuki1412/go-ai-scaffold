# Effective Go 知识图谱（三源融合版）

> **融合来源**（按权威度排序）：
> 1. Go 官方 [Effective Go](https://golang.google.cn/doc/effective_go) —— 精读全文（2009 年撰写、含后续修订，如 Go 1.22 循环变量、Go 1.26 `new(expr)` 说明）；
> 2. 《Go专家编程》（RainbowMango / GoExpertProgramming，2019，Go ~1.12 时代）—— 深入运行时实现原理；
> 3. 《Mastering Go》中文第一版（Mihalis Tsoukalos，2017，Go 1.9/1.10）—— 工程实战与系统/网络编程。
>
> **现代基准：Go 1.25**（本仓库 `go.mod` 版本）。两本书成书较早，凡与 Go 1.25 现状不符之处以「⚠️1.25」标注。
> **行内标记**：〔EG〕〔GEP〕〔MG〕标示知识点主要出处；💡 = 值得固化的工程惯用法；⚠️ = 易踩坑。

---

## 目录

1. [格式与注释](#1-格式与注释)
2. [命名](#2-命名)
3. [分号与控制流](#3-分号与控制流)
4. [函数](#4-函数)
5. [数据](#5-数据)
6. [初始化与包](#6-初始化与包)
7. [方法与接口](#7-方法与接口)
8. [嵌入与组合](#8-嵌入与组合)
9. [并发](#9-并发)
10. [错误处理](#10-错误处理)
11. [内存与性能](#11-内存与性能)
12. [反射与 unsafe](#12-反射与-unsafe)
13. [测试、基准与工具链](#13-测试基准与工具链)
14. [文件与系统编程](#14-文件与系统编程)
15. [网络与 HTTP](#15-网络与-http)
16. [定时器与资源泄露](#16-定时器与资源泄露)
17. [语法糖：`:=` 与 `...`](#17-语法糖--与-)
18. [Go 1.9/1.12 → Go 1.25 演进速查](#18-go-19112--go-125-演进速查)
19. [本项目（go-ai-scaffold）对照审计清单](#19-本项目go-ai-scaffold对照审计清单)

---

## 1. 格式与注释

### 1.1 gofmt：把格式问题交给机器〔EG〕

- `gofmt`（包级命令 `go fmt`）按标准风格输出代码，**注释也会自动对齐成列**。遇到新排版问题先跑 gofmt；结果不合理就重排程序结构，不要绕过工具。
- 缩进用 **tab**（迫不得已才用空格）；**没有行长限制**，太长就换行并多缩进一层。
- 控制结构语法上不带括号；运算符优先级层次更短更清晰（`x<<8 + y<<16` 中移位优先于加法，按间距即可读出结合性）。

### 1.2 注释与 godoc〔EG〕

- 行注释 `//` 是常态；块注释 `/* */` 用于包注释、表达式内部或临时屏蔽大段代码。
- **出现在顶层声明之前、中间无空行的注释即该声明的文档注释**，是包/命令的主要文档。
- 文档注释以标识符名开头（`// Foo does ...`），godoc 会按此渲染；包注释在每个包中只需一份（任意文件均可，习惯放 doc.go）。

---

## 2. 命名〔EG〕

### 2.1 核心规则

- **首字母是否大写决定包外可见性**——名字有语义效果。多词用 `MixedCaps` / `mixedCaps`，不用下划线〔EG〕。
- 包名：**小写、单个单词、简短达意**，不起下划线或 mixedCaps；包名取源目录名，不必全局唯一。导出名利用包名上下文避免重复：`bufio.Reader` 而非 `BufReader`；ring 包唯一类型的构造函数直接叫 `ring.New`〔EG〕。
- 长名不一定更可读：`once.Do(setup)` 好过 `once.DoOrWaitUntilDone(setup)`。一条好注释常胜过超长名字〔EG〕。
- 避免 `import .` 写法〔EG〕。

### 2.2 Getter / 接口名

- Go 不自动生成 getter/setter；自己提供时 **getter 不加 Get 前缀**：字段 `owner` 的 getter 叫 `Owner`，setter 叫 `SetOwner`〔EG〕。
- 单方法接口用「方法名 + -er 后缀」构成施动者名词：`Reader`、`Writer`、`Formatter`、`CloseNotifier`〔EG〕。
- `Read/Write/Close/Flush/String` 等有规范签名与含义：签名或语义不同就不要用这些名字；语义相同就应同名同签名（叫 `String` 而非 `ToString`）〔EG〕。

---

## 3. 分号与控制流

### 3.1 分号规则〔EG〕

- 源码基本不写分号，由词法器**在换行前最后一个 token 是「标识符 / 字面量 / `break continue fallthrough return ++ -- ) }`」时自动插入**。
- 重要后果：**控制结构的左花括号不能放到下一行**（否则花括号前被插分号）。
- 分号仅用于：for 子句三段分隔、同行多语句。

### 3.2 if / for / switch / 类型 switch〔EG〕

- `if` 与 `switch` 可带前置初始化语句：`if err := file.Chmod(0664); err != nil { ... }`。
- **if 体以 break/continue/goto/return 结尾时省略多余 else**，让成功路径沿页面下行、逐个排除错误分支。
- `:=` 重声明规则：同作用域内已声明的变量可在多值 `:=` 中再次出现（此时仅重新赋值），前提是①同作用域 ②类型可赋值 ③**至少还引入了一个新变量**〔EG〕。
- for 三形态：`for init; cond; post {}`、`for cond {}`（即 while）、`for {}`（无限循环）。range 可遍历数组、切片、字符串、map、channel〔EG〕。
- **range 字符串按 UTF-8 解码产出 rune**；非法编码消耗一字节并产出 U+FFFD〔EG〕。
- Go **无逗号运算符**；`++`/`--` 是语句不是表达式。多变量推进用并行赋值：`for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1`〔EG〕。
- switch：case 不必是常量甚至不必是整数，从上到下求值；**无表达式 switch（`switch {}`）相当于 `switch true`**，可把 if-else-if 链改写得更清晰；**无自动 fallthrough**；case 可用逗号列表（`case ' ', '?', '&':`）〔EG〕。
- break 跳出 switch；要跳出外层循环需**给循环加标签** `break Loop`；continue 也接受标签但只作用于循环〔EG〕。
- 类型 switch：`switch v := x.(type) { case int: ... }`，v 在各分支中自动具对应具体类型〔EG〕。

### 3.3 range 底层与坑〔GEP〕

- range 由编译器改写为 C 风格 for：**slice/数组的循环次数在循环开始前由 len 确定**——循环中 append 改变长度不影响本轮循环次数〔GEP〕。
- range map 遍历顺序随机；遍历中插入的新键不保证被遍历到〔GEP〕。
- range channel：`for v := range ch` 持续读，**channel 关闭时退出**；写方不关闭则永久阻塞〔GEP〕。
- 性能：range 大切片且不需 value 时只取 index 再 `s[i]` 引用，避免每次迭代的 value 拷贝〔GEP〕。
- ⚠️1.25：**Go 1.22 起 for-range 循环变量每次迭代是新建变量**，旧版「闭包捕获共享循环变量」的经典坑已由语言修复；Go 1.23 起 range 支持函数迭代器（range-over-func），Go 1.22 起支持 `for i := range n`（range over int）。

---

## 4. 函数

### 4.1 多返回值与命名返回值〔EG〕

- 多返回值改善 C 的带内错误返回与传地址修改参数两大笨拙用法（`Write(p []byte) (n int, err error)`）。
- 命名返回值：函数开始时初始化为零值，裸 `return` 返回当前值；**名字即文档**（标明哪个是值、哪个是位置）〔EG〕。
- ⚠️ 命名返回值配合 defer 可改写最终返回值（见 4.3）；函数体内 `min := ...` 会遮蔽命名返回值，裸 return 时返回零值〔MG〕。

### 4.2 可变参函数〔GEP〕〔EG〕

- `...T` 必须位于参数列表尾部；函数内部作为切片 `[]T` 处理；调用时不传则为 **nil 切片**；需异质参数时声明 `...any`〔GEP〕。
- 三种调用方式：不传 / 传多个值（编译器打包）/ **传切片 `s...` 展开**〔GEP〕。
- ⚠️ **`s...` 展开不拷贝**：传入切片与函数内部切片共享底层数组，函数内修改元素可能影响调用方；反向也不可靠（append 扩容后脱钩）。把可变参切片当只读入参；要改先克隆〔GEP〕。

### 4.3 defer〔EG〕〔GEP〕〔MG〕

- defer 把函数调用安排在**外围函数返回之前**执行（函数级、非块级），适合「无论走哪条返回路径都要释放资源」——解锁互斥、关闭文件是经典用法。优点：①新增 return 路径也安全 ②释放语句紧挨获取语句〔EG〕。
- **官方三原则**〔GEP〕：
  1. **延迟函数的参数在 defer 语句出现时确定**（值拷贝固化；指针参数拷贝的是地址）；
  2. **按 LIFO 顺序执行**（后 defer 的先执行，适配「资源依赖 A→B→C，释放反向 C→B→A」）；
  3. **延迟函数可操作主函数的具名返回值**——因为 `return` 不是原子操作：`设置返回值 → 执行 defer → ret 跳转`。
- 经典对照〔MG〕：循环中 `defer fmt.Print(i)` 输出 1 2 3（参数固化）；`defer func(){ fmt.Print(i) }()` 捕获变量本身（Go 1.22 前输出 0 0 0）；`defer func(n int){ fmt.Print(n) }(i)` 输出 1 2 3。💡 **给 defer 的闭包显式传参**是可读性最佳实践。
- 每申请到一个需释放的资源，**立即写一个 defer**；defer 在 panic 时也会执行（资源清理兜底）〔GEP〕。
- ⚠️1.25：Go 1.14 起 defer 采用 open-coded 实现（非循环内、个数 ≤8 等条件），开销近乎为零——「defer 昂贵」的旧观念不成立；但**循环内大量 defer 仍会累积**，需注意〔GEP〕。

### 4.4 闭包〔MG〕

- 闭包捕获的是**变量本身（引用）**，跨调用持续保留状态；每次调用外层函数生成全新、互不干扰的闭包环境〔MG〕。
- 闭包陷阱清单〔MG〕：①状态滞留（计数器跨调用保留）；②函数变量可被重新赋值指向别的函数（「可修改的匿名函数变量是错误的根源」）；③循环变量捕获（⚠️1.22 起已由语言修复）；④「谁在何时改了这个变量」难以追踪——控制闭包的生命周期与捕获集合。
- 匿名函数的适用边界：**小而聚焦**；不聚焦就升级为普通函数〔MG〕。
- 闭包引用的局部变量会逃逸到堆（`moved to heap`），见 §11.2〔GEP〕。

---

## 5. 数据

### 5.1 new / make / 零值可用〔EG〕

- `new(T)` 分配**零值**内存返回 `*T`，只清零不初始化；`make(T, args)` **只用于 slice/map/channel**，返回**已初始化**的 `T`（非指针）——这三种类型底层是引用结构，必须先初始化才能用〔EG〕。
- `new([]int)` 返回指向 nil 切片的指针，几乎无用；惯用 `v := make([]int, 100)`〔EG〕。⚠️1.25 补充：Go 1.26 起将支持 `new(expr)` 带初值。
- 💡 **让类型的零值可直接使用**：`bytes.Buffer` 零值即空缓冲、`sync.Mutex` 零值即未锁定互斥；零值可用性是**可传递的**（嵌入后组合类型零值也能用）〔EG〕。
- 复合字面量：`&File{fd, name, nil, 0}`——**返回局部变量地址完全合法**；字段 `名:值` 显式标注时可任意顺序、缺省取零值；`&File{}` 等价 `new(File)`；复合字面量同样适用于数组/切片/map〔EG〕。

### 5.2 数组〔EG〕

- **数组是值**：赋值与传参拷贝全部元素；**大小是类型的一部分**（`[10]int` ≠ `[20]int`）。
- 要 C 式效率可传数组指针，但不惯用——**用切片代替**；只有「元素数量非常确定」（如变换矩阵）才用数组〔EG〕〔MG〕。

### 5.3 slice〔EG〕〔GEP〕〔MG〕

- 底层三字段结构体〔GEP〕：
  ```go
  type slice struct {
      array unsafe.Pointer // 指向底层数组
      len   int
      cap   int
  }
  ```
- 函数传参只拷贝切片头（指针+len+cap），**不拷贝底层数组**：函数内**修改元素**对调用方可见，但 **append 触发扩容后的新切片头不会自动传回**——须返回新切片或传 `*[]T`〔GEP〕〔MG〕⚠️初学者最常踩的坑。
- 三索引切片 `s[low:high:max]`：新切片容量为 `max-low`（以 max 截断，可与原数组隔离）〔GEP〕。
- append：容量够则原地追加；不够则**重新分配更大内存 + 拷贝旧数据**，必须接收返回值〔EG〕。扩容策略：书口径「<1024 翻倍、≥1024 ×1.25」⚠️1.25：Go 1.18 起改为渐进公式（小容量翻倍、约 >256 后按 ≈1.25 平滑增长并按 size class 取整），别背数字，记住「增长有成本、预估容量最划算」〔GEP〕〔MG〕。
- copy：拷贝个数 = **min(len(dst), len(src))**，静默截断不报错，不扩容；返回实际拷贝个数〔EG〕〔MG〕。
- **re-slicing 两大坑**〔MG〕：①共享底层数组——改小切片会动大切片；②从大切片切出小切片会**整体拖住**大数组内存不释放。需要独立数据用 copy / `slices.Clone`。
- 二维切片两种分配：逐行 make（行可伸缩）或一次性大数组再切（行固定更高效）〔EG〕。
- 💡 预分配容量（make 第三参数）避免反复扩容——GEP 基准实测预分配比不预分配快 3 倍+〔GEP〕。

### 5.4 map〔EG〕〔GEP〕〔MG〕

- 键可为任何定义了相等运算的类型（含结构体、数组）；**切片不能做键**〔EG〕。
- map 持有底层结构引用：函数内改动对调用方可见〔EG〕。
- **查不存在的键返回零值**；需区分「零值」与「不存在」用 **comma-ok**：`v, ok := m[k]`；删除 `delete(m, k)`（键不存在也安全）〔EG〕。
- ⚠️ **nil map 读得零值、写则 panic**（`assignment to entry in nil map`）〔MG〕。
- ⚠️ **map 非并发安全**：并发读写触发不可 recover 的 `fatal error: concurrent map writes`——需 `sync.Mutex` / `sync.RWMutex` / `sync.Map` 保护〔GEP〕〔MG〕。
- 底层（Go ≤1.23）〔GEP〕：哈希表 + `2^B` 个桶、每桶 8 槽、tophash 高 8 位快速比对、overflow 链（链地址法）；**负载因子 6.5** 或 overflow 桶过多触发扩容；**渐进式搬迁**（增量扩容）+ 等量扩容（压缩 overflow）。⚠️1.25：**Go 1.24 起 map 底层改为 Swiss Table**（group + 控制字节 + SIMD 探测），桶结构描述已过时，但「负载因子 6.5、渐进扩容、非并发安全」的结论仍成立。
- 💡 预估容量 `make(map[K]V, n)` 减少扩容开销〔GEP〕。

### 5.5 string / rune / byte〔EG〕〔GEP〕〔MG〕

- string 底层 `{str unsafe.Pointer; len int}`，是**8-bit 字节的集合（通常 UTF-8）、不可变、可为空但不会是 nil**〔GEP〕。
- `[]byte ↔ string` 转换默认**一次内存拷贝**；编译器免拷贝优化的三类临时场景：map 查找 `m[string(b)]`、拼接 `"<"+string(b)+">"`、比较 `string(b)=="foo"`〔GEP〕。
- 取舍〔GEP〕：string 擅长比较、不需 nil 语义；[]byte 擅长修改、用 nil 表达含义、切片操作。热路径避免来回转换（⚠️1.25：零拷贝可用 `unsafe.String`/`unsafe.Slice`，见 §12）。
- **len(s) 数的是字节数不是字符数**（`len("€£³") == 7`）〔MG〕。rune = `int32` 的 **Unicode 码点**，字面量用单引号 `'€'`〔MG〕。
- 两种遍历语义：下标按 **byte**，range 按 **rune**（返回字节偏移 + 码点）〔MG〕。
- 一句话总结〔MG〕：**字节是 8bit 存储单元；rune 是一个 Unicode 码点；字符串是字节序列，按 UTF-8 解码后构成 rune 序列**。
- 💡1.25：逐段拼字符串用 `strings.Builder` 别用 `+=`；`strings.Cut` 是最优雅的分隔符解析。

### 5.6 const 与 iota〔EG〕〔GEP〕〔MG〕

- 常量编译期创建，只能是数字/rune/字符串/布尔；初始化必须是编译期可求值的常量表达式（`1<<3` 可以，`math.Sin(...)` 不行）〔EG〕。
- **iota 的准确规则：它是 const 声明块的行索引（下标从 0 开始）**；同一行内多次使用值相同（不递增）；**后续行无表达式则继承上一行的表达式**〔GEP〕。
- 惯用法：`_ = iota` 跳行 + `KB ByteSize = 1 << (10 * iota)` 自动沿用〔EG〕；`1 << iota` 定义位掩码（源自 sync.Mutex 源码）〔GEP〕。
- **无类型常量 vs 有类型常量**：`const s1 = 123` 可自由参与各数值运算；`const s2 float64 = 123` 则受类型约束。💡 除非必要不给常量标类型〔MG〕。

### 5.7 struct 与 Tag〔GEP〕〔MG〕

- 结构体赋值是**深拷贝**；字段顺序是类型的一部分；`var s T` 字段自动零值〔MG〕。
- 字面量两种写法：位置式（须全字段按序）与**键值式（推荐，免记顺序可缺省）**〔MG〕；`NewXxx()` 构造函数集中校验是标准惯例〔MG〕〔EG〕。
- **Tag 语法**：以空格分隔的 `key:"value"` 序列；value 必须双引号包裹；**冒号前后不能有空格**——写错不报编译错误，只会反射时静默取不到〔GEP〕⚠️。
- Tag 服务于反射场景：`encoding/json`、ORM 字段映射均基于此；Go 1.7+ 有 `StructTag.Lookup`（可区分「不存在」与「空值」）〔GEP〕。

---

## 6. 初始化与包

- 变量初始化器可以是运行时表达式（`os.Getenv("HOME")`）〔EG〕。
- `init()`：每个源文件可有一个或多个；**在包内所有变量初始化器求值之后、且仅在所有被导入包初始化完成后**执行；常用于无法用声明表达的初始化或程序状态校验/修复（如检查 `$USER`、注册 flag）〔EG〕。
- 包的可见性：**标识符首字母大写 = 导出**，函数/类型/常量/字段一视同仁；`internal/` 目录提供「仅模块内可见」的更强封装〔MG〕。
- 包设计：小而聚焦的包、从使用者视角命名（见 §2.1）；**接口定义在消费方一侧**（见 §7.3）〔MG〕。
- ⚠️ 别在 init 里做重活（网络/文件 IO）——启动顺序不可控且难测试〔通用工程共识〕。

---

## 7. 方法与接口

### 7.1 接收者规则〔EG〕

- 方法可定义在任何命名类型上（指针和接口除外），接收者不必是结构体〔EG〕。
- **核心规则：值方法可通过值和指针调用；指针方法只能通过指针调用**。指针方法可修改接收者；在值上调用时方法收到的只是副本、修改会被丢弃，语言直接禁止〔EG〕。
- 便利例外：值可寻址时编译器自动插入取址（`b.Write` → `(&b).Write`）〔EG〕。
- 接收者命名：普通变量名、通常单字母，**不用 this/self**〔MG〕。
- 选择建议〔通用〕：需要修改接收者 / 结构体较大 / 类型一致性（该类型其他方法已用指针）→ 指针接收者；小值类型、map/chan/函数类型 → 值接收者。

### 7.2 接口：隐式实现〔EG〕

- 接口指定行为——「能做这件事」即可在此使用；一个类型可实现多个接口（如 Sequence 同时满足 `sort.Interface` 和自带 `String()`）〔EG〕。
- 最常见的是**单方法或双方法接口**，通常以方法名命名（`io.Writer`）；「力量源于简单」〔EG〕〔MG〕。

### 7.3 接口设计惯用法〔MG〕

- 💡 **接口应定义在消费方一侧**：「若发现接口及其实现定义在同一个 Go 包里，可能用错了接口」。
- 只为实现某接口而存在、且不导出接口之外方法的类型，**只导出接口即可**；**构造器应返回接口值**（`crc32.NewIEEE` 与 `adler32.New` 都返回 `hash.Hash32`，替换算法只改构造调用）〔EG〕。
- 编译期断言实现〔EG〕：`var _ json.Marshaler = (*RawMessage)(nil)`——接口变更时包将编译失败。
- 运行时检查实现：`if _, ok := v.(json.Marshaler); ok`〔EG〕。

### 7.4 类型断言与类型 switch〔EG〕

- 单类型断言 `v.(T)`：结果是静态类型为 T 的新值；**单值形式失败即 panic**；**双值（comma-ok）形式安全**：`s, ok := v.(string)`，失败时 s 为零值〔EG〕〔MG〕。
- 类型 switch 本质是逐 case 的转换；惯用法是复用原变量名 `switch v := x.(type)`〔EG〕。
- 惯用法：处理 any 数据（JSON、`list.Element.Value`）必用；能用窄接口就别到处断言〔MG〕。

### 7.5 Go 的「OOP」思想〔MG〕

- Go 无继承，支持**组合**；接口提供多态——可**模拟** OOP 但不是 OOP 语言；有意禁止深类型层次。
- **嵌入 ≠ 继承**：second 嵌入 first、两者都有 `shared()` 时，`first{}.F()` 内部调用的 `shared()` **永远解析到 first 版本**——组合不提供动态派发；要「重写」语义必须用接口〔MG〕。

---

## 8. 嵌入与组合

- 接口内嵌：`io.ReadWriter = Reader + Writer` 的并集；只有接口可内嵌于接口〔EG〕。
- 结构体内嵌（不写字段名）：**内嵌类型的方法免费获得**，同时满足多个接口，免手写转发方法〔EG〕。如 `bufio.ReadWriter` 内嵌 `*Reader` 与 `*Writer`。
- **与子类化的关键区别：方法被调用时接收者是内层类型，不是外层类型**〔EG〕。
- `type Job struct { Command string; *log.Logger }`：Job 获得 Logger 的 Print/Printf 等方法；需直接引用时用类型名作字段名 `job.Logger`〔EG〕。
- 命名冲突规则：①浅层遮蔽深层；②同一层级同名通常是错误，但**若该重名在类型定义之外从未被引用则没问题**〔EG〕。

---

## 9. 并发

### 9.1 并发 vs 并行〔EG〕〔MG〕

- 口号：**「不要通过共享内存来通信，而要通过通信来共享内存」**。共享值在 channel 上传递，任一时刻只有一个 goroutine 拥有该值，数据竞争在设计上不可能发生〔EG〕。
- 此法不可走极端——引用计数等场景 mutex 更合适；但作为高层方法，channel 更易写出清晰正确的程序〔EG〕。
- **并发（结构化为独立执行组件）≠ 并行（多 CPU 同时执行）**："Go is a concurrent language, not a parallel one"〔EG〕；「并发性优于并行性」——先做有效的并发设计，并行能力自然到来〔MG〕。

### 9.2 goroutine 与 GMP 调度〔EG〕〔GEP〕〔MG〕

- goroutine 与其他 goroutine 在同一地址空间并发执行的函数，开销仅比栈分配略高；栈初始很小（2KB）按需增长；多路复用到多个 OS 线程〔EG〕。
- GMP 三对象〔GEP〕：**G**（goroutine）、**M**（OS 线程）、**P**（逻辑处理器，持有运行 G 所需资源与本地队列）；**M 必须拥有 P 才能执行 G**；P 默认 = CPU 核数（⚠️1.25：Linux 容器下默认取 cgroup CPU 配额，向上取整）。
- 三种调度策略〔GEP〕：①**队列轮转**（P 本地队列分时轮转 + 周期性查看全局队列防饿死）；②**系统调用 hand-off**（M0 陷入系统调用时释放 P 给空闲 M1 继续跑队列中的 G；G0 返回后抢不到 P 则进全局队列）；③**work stealing**（空闲 P 从其他 P 偷取约一半 G）。
- `runtime.GOMAXPROCS(n)` 设置 P 数；调大不会让 CPU 密集程序更快〔MG〕。
- ⚠️1.25：书中「IO 密集就调大 GOMAXPROCS」是 netpoller 完善前的旧经验——现代 Go 网络阻塞的 G 由 netpoller 管理不占 M；容器环境反而常用 automaxprocs 收紧。**Go 1.14 起基于信号（SIGURG）的异步抢占**解决了紧密循环饿死问题〔GEP〕。
- 性能忠告〔MG〕：goroutine 不是越多越好——更多 goroutine 意味着调度器更多工作。
- `go` 语句立即返回；**无法自然控制 goroutine 的执行顺序**；`time.Sleep` 等待 goroutine 是不可靠的原始手段（main 返回即全灭）——用 WaitGroup/channel/context〔MG〕。

### 9.3 channel〔EG〕〔GEP〕〔MG〕

- `make(chan T)` 无缓冲（默认）/ `make(chan T, n)` 缓冲 n；channel 与 map 一样是对底层结构的**引用**〔EG〕。
- **无缓冲 channel 将通信与同步合一**：发送方阻塞直至接收方收到；缓冲 channel 发送方阻塞直至值拷入缓冲〔EG〕。缓冲 channel 可作**信号量限流**〔EG〕。
- 底层 hchan = 环形队列 + 类型信息 + sendq/recvq 双等待队列 + 锁〔GEP〕；单向 channel 本质不存在，只是参数类型约束（`chan<- int` 只写 / `<-chan int` 只读），**给函数参数标注方向是重要的 API 自文档化**〔GEP〕〔MG〕。
- **操作语义矩阵**（必背）〔GEP〕〔MG〕：

  | 操作 | nil channel | 已关闭 | 正常 |
  |---|---|---|---|
  | 发送 `ch <-` | 永久阻塞 | **panic** | 未满写入；否则阻塞 |
  | 接收 `<-ch` | 永久阻塞 | 先读完缓冲，后立即返回零值（ok=false） | 有值取出；否则阻塞 |
  | `close(ch)` | **panic** | **panic**（重复关闭） | 成功；唤醒所有等待者 |

- **close 是广播**：同时解除所有阻塞在该通道上的接收者；发送值只能唤醒一个〔MG〕。
- channel of channel〔EG〕：channel 是一等值可传递；`Request` 内嵌 `resultChan` 让客户端自带回收结果的路径——限速、并行、非阻塞 RPC 的骨架，全程无 mutex。
- 生产者-消费者/管道（pipeline）：前级输出为后级输入；**忘记 close 通道 → `for range` 永久阻塞**；管道的价值：数据流式处理、省内存、设计简化〔MG〕。

### 9.4 select〔EG〕〔GEP〕〔MG〕

- select 是语言层面的多路 IO 复用：一个 select 监听多个 channel；**不需要 default**；**各 case 同时被检查、多个就绪时随机（公平）选一个**〔MG〕。
- 底层〔GEP〕：每个 case 一个 `scase`；`pollorder` 洗牌实现随机检测、`lockorder` 按地址排序去重加锁防死锁。
- 空 select `select{}` 永久阻塞（触发死锁检测 panic）；**已关闭的 channel 也是「可读」的**——读 case 可能落入，务必 `v, ok := <-ch` 检查〔GEP〕。
- `select + default` 实现非阻塞发送/接收（缓冲满/空走 default）〔MG〕⚠️：default 里的 `break` 只跳出 select 不跳出外层 for，须用标签或 return。
- 💡 **nil 通道禁用分支**：把某 case 的 channel 赋成 nil，该分支「永不就绪」等于从 select 中摘除——状态机式 select 的经典惯用法〔MG〕。
- `time.After` 当「软 default」实现超时〔MG〕；⚠️1.25：Go 1.23 前长循环里每轮新建 `time.After` 是定时器泄漏隐患，1.23 起未触发的 Timer 可被 GC 及时回收。

### 9.5 WaitGroup〔GEP〕〔MG〕

- 信号量机制：`Add(n)` 置计数 → 每 goroutine `defer Done()` → `Wait()` 阻塞至归零〔MG〕。
- 铁律〔MG〕〔GEP〕：**`Add(1)` 必须在 `go` 语句之前**（放 goroutine 内部与 Wait 构成竞争）；**Done 一律 defer**；Add 与 Done 数量必须配对。
- 后果〔MG〕：Add > Done → `Wait()` 永久阻塞（全体睡眠时 runtime 报 deadlock，否则**静默卡死**更难排查）；Done > Add → `panic: sync: negative WaitGroup counter`。
- WaitGroup 只保证「全部完成」，**不保证顺序**〔MG〕。
- 适用对比〔GEP〕：Channel 控制简单直接；WaitGroup 计数动态可调；**Context 对多级（树状）goroutine 控制力最强**。⚠️1.25：现代常配 `golang.org/x/sync/errgroup`（带错误传播与 SetLimit）。

### 9.6 Mutex / RWMutex〔GEP〕〔MG〕

- 互斥 ≈ 容量为 1 的缓冲通道；**识别临界区是程序员的核心任务**〔MG〕。
- 💡 **拿到锁立即 `defer mu.Unlock()`**——一旦中间提前 return 或 panic，锁永远不释放；忘记 Unlock：单次调用一切正常（隐蔽），两个 goroutine 即「全员睡眠死锁」，且系统里还有活跃 goroutine 时会**无声挂死**〔MG〕。
- 纪律〔MG〕：同一把锁的两个临界区不能嵌套；**不惜代价避免跨函数传播互斥体**；重复 Unlock 会 panic。
- Mutex 底层〔GEP〕：state 位段（Locked/Woken/Starving/Waiter 计数）+ 信号量；**自旋条件**（最多 4 次、多核、P 空闲、本地队列空）；**Normal/Starving 双模式**——等待超 1ms 切饥饿模式（不自旋、释放必交接）保障尾延迟。
- RWMutex〔GEP〕〔MG〕：多读并发 + 单写独占；`w Mutex + 双信号量 + readerCount(±2^30 技巧) + readerWait`；写者把读流切成两段防饿死。**适用：读远多于写且读临界区较耗时**（书实测快一倍+）；读临界区极短时维护成本反超 Mutex——先测再选。RLock 区间内绝不修改共享变量；**不可重入**。
- 💡 **监视器 goroutine**：让单个 goroutine 独占拥有共享数据，其他 goroutine 通过 channel 读写（`readValue <- value` 的 select）——不用锁天然无竞争，适合长生命周期服务〔MG〕。
- 并发编程元原则〔MG〕：**除非万不得已，避免共享**——共享数据是并发 bug 的根本原因。

### 9.7 context〔GEP〕〔MG〕

- 四方法接口：`Deadline() / Done() / Err() / Value()`；使用者**无需自己实现**，用 `WithCancel / WithTimeout / WithDeadline / WithValue` 从父派生〔MG〕。
- 树形传播规则〔GEP〕：**取消自上而下级联**（children 递归 cancel）；**Value 查找自下而上回溯父链**；`Done()` 返回的 channel 在自己被 cancel **或父链任一节点关闭**时关闭。
- `defer cancel()` 惯用法；`Err()` 给出原因：`context canceled` / `context deadline exceeded`〔MG〕。
- ⚠️ **单独 `WithValue(Background(), ...)` 派生的 context 不可取消**（`<-ctx.Done()` 永远不返回）；需要可取消时必须有可 cancel 的父节点〔GEP〕。
- 官方惯例：context 作为函数**第一个参数**；不要存入结构体；`go vet` 的 lostcancel 检查未调用的 cancel〔GEP〕。
- ⚠️1.25：Go 1.20+ 新增 `WithCancelCause/WithTimeoutCause/Cause`、`AfterFunc`、`WithoutCancel`；取代了废弃的 `Transport.CancelRequest`（现代写法 `http.NewRequestWithContext`）〔MG〕。

### 9.8 常用并发模式集

- **超时两法**〔MG〕：①`select { case <-结果通道; case <-time.After(d) }`——简单直接，超时后原 goroutine 仍在跑（注意泄漏）；②封装 `timeout(w *WaitGroup, t)` 对「整组等待」加超时。现代首选 **context** 统一表达超时与取消并贯穿调用链。
- **信号通道**〔MG〕：只发信号不传数据的通道，类型选 `chan struct{}`——从类型上杜绝误发数据且零字节开销。
- **执行顺序**〔MG〕：`chan struct{}` 信号链 + `close` 广播点火（close x → A → close y → B → ...）；⚠️ 会 close 通道的函数只能启动一个实例（重复 close panic）。
- **工作池 worker pool**〔MG〕：缓冲通道 = 任务/结果队列（缓冲容量即并发上限）；worker 以 `for range` 消费、**任务通道 close 驱动退出**；WaitGroup 管 worker 生命周期；信号通道收尾消费者。⚠️1.25：生态常用 `errgroup` 替代手写池。
- **泄漏缓冲池 leaky buffer**〔EG〕：`select + default` 从 freeList 非阻塞取/还 buffer，满了交给 GC——几行实现 free list。
- **并发服务两板斧**〔MG〕：HTTP 每请求一 goroutine；TCP `for { c := l.Accept(); go handleConnection(c) }`。
- ⚠️ goroutine 内部错误应 `return` 结束本 goroutine，**绝不该 `os.Exit` 带崩整个进程**〔MG〕。

### 9.9 竞争状态与 race detector〔MG〕

- 数据竞争定义：两个及以上指令访问同一内存地址且至少一个是写。
- **「跑一次看结果」无法发现竞争**；多次运行结果不一致是竞争的信号；**`-race`（go run/build/test）是权威判据**——编译出记录所有共享访问与同步事件的版本，打印 `WARNING: DATA RACE` 报告（含双方 goroutine 栈）。
- 两大高频事故〔MG〕：循环变量捕获（⚠️1.22 起语言层面消灭）与**并发写 map**（不加锁直接 fatal，不可 recover）。
- ⚠️1.25：race detector 基于 TSAN v3（Go 1.19+），有数倍内存/CPU 开销——用于测试与 CI，不用于生产构建；`go test -race` 应纳入 CI 默认项。

---

## 10. 错误处理

### 10.1 error 接口与规范〔EG〕

- `type error interface { Error() string }`；多值返回使「正常返回值 + 详细错误」很自然〔EG〕。
- 错误字符串应表明来源，如以操作名或包名作前缀（`image: unknown format`）〔EG〕。
- 库作者可用更丰富的底层模型（如 `*os.PathError` 含 Op/Path/Err），**需要细节的调用方用类型断言提取**：`if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOSPC { ... }`〔EG〕。
- 💡1.25 工程规范（Go Code Review Comments）：错误字符串**小写、无标点、不首字母大写**（大写开头留给专有名词与缩写）；包装错误用 `fmt.Errorf("...: %w", err)` + `errors.Is/As` 判等/取因〔MG 笔记补充〕。
- ⚠️ **永远不要用 `_` 丢弃错误**："Always check error returns; they're provided for a reason."〔EG〕。

### 10.2 panic / recover 的边界〔EG〕〔MG〕

- 常规报错返回 error；**panic 仅用于程序确实无法继续**的情况（不变量破坏、初始化失败如 `$USER` 为空、标记「不可能发生」的分支如百万次迭代不收敛）〔EG〕。
- 真实库函数应避免 panic——能绕过就让程序继续跑〔EG〕。
- recover 语义〔EG〕：panic 立即停止当前函数并退栈，沿途执行 deferred 函数，退到 goroutine 栈顶程序死亡；**recover 只在 deferred 函数内直接调用才有效**，它停止退栈并返回传给 panic 的值。
- 典型应用：服务器中让失败的 goroutine 干净退出而不殃及其他（`safelyDo` 在 defer 闭包中 recover 并记日志）〔EG〕。
- **包内 panic 转 error 惯用法**〔EG〕：解析器内部用 `re.error("...")` 方法 panic 简化深层报错，顶层 `Parse` defer recover 转成 error 返回——**不把 panic 暴露给客户端**；re-panic 惯用法保留崩溃根因。
- 💡 工程共识〔MG〕：这正是「service 层 panic 业务错误、middleware 层统一 Recover 转错误响应」框架设计的理论出处（本项目 `middleware.Recover` + `panic(exception.New(...))` 约定同源）。**recover 不能跨 goroutine**；panic 之后的代码不执行，「事后清理」必须写在 defer 里。
- `log.Fatal` = 打日志 + `os.Exit(1)`，**已注册的 defer 一律不执行**，只用于不可恢复的启动错误；`log.Panic` = 打日志 + panic，可被上层 recover〔MG〕。

---

## 11. 内存与性能

### 11.1 内存分配（tcmalloc 思想）〔GEP〕

- 三级架构：**mcache（P 私有，无锁）→ mcentral（全局按 class，加锁）→ mheap（全局唯一）**；span 是基本管理单位（1~n 个连续 8KB 页），按 67 个 size class（8B~32KB）切分；>32KB 大对象走 mheap；<16B 无指针走 Tiny 合并分配〔GEP〕。
- mcache 按 `[67*2]` 分组：**scan（含指针）/ noscan（不含指针）二分**——不含指针的 span 无需 GC 扫描，**结构体里少放指针有利于 GC**〔GEP〕。
- ⚠️1.25：Go 1.11 起堆为稀疏地址空间（按 64MB arena 映射，「512GB 三段布局」说法过时）；mcentral 的 nonempty/empty 链表已被 spanSet 取代；新增 `GOMEMLIMIT` 软内存上限（1.18+）。核心思想（P 私有缓存无锁快路径 + 全局慢路径）不变。

### 11.2 逃逸分析〔GEP〕

- 编译器决定分配位置：**函数外部没有引用 → 优先栈；存在引用 → 必定堆**（过大或长度不定也可能堆）。
- 四大逃逸场景〔GEP〕：①**指针逃逸**（返回局部变量指针——这也是「返回局部变量地址安全」的原因）；②**栈空间不足**（大对象/变长 make）；③**动态类型逃逸**（`fmt.Println(a ...any)` 等 interface 参数——热路径慎用）；④**闭包引用**（`moved to heap: a`）。
- 判定：`go build -gcflags="-m"`，关注 `escapes to heap` / `moved to heap` / `does not escape` / `leaking param`。
- 💡 **传指针不一定比传值高效**：小对象传指针会引发逃逸（堆分配 + GC 负担），反而更慢；小对象传值、大对象才传指针〔GEP〕。

### 11.3 垃圾回收〔GEP〕〔MG〕

- Go = **并发三色标记清除 + 写屏障，非分代、非压缩**，设计目标**低延迟而非极致吞吐**——调优手段主要是**减少分配**，而非「选 GC」〔MG〕。
- 三色语义：白（候选回收）/ 灰（已标记待扫描，工作队列）/ 黑（扫描完毕，**不允许持有指向白对象的指针**——算法不变式）；写屏障在 mutator 改指针时把被指对象标灰以维持不变式〔GEP〕〔MG〕。
- 触发时机〔GEP〕：①分配量达阈值（`阈值 = 上次 GC 后堆 × GOGC/100`，**GOGC 默认 100** 即堆翻倍触发）；②**最长 2 分钟**强制一次；③`runtime.GC()` 手动（**会阻塞调用者**，别随手调）。
- GC 性能与**对象数量**负相关〔GEP〕：减少分配个数（池化复用、大对象组合小对象）；警惕逃逸带来的隐式分配。
- 观测〔MG〕：`GODEBUG=gctrace=1`（三个数字 = GC 前/后堆 + 存活堆）；`runtime.ReadMemStats`（⚠️会 STW，生产别高频调，用 `runtime/metrics`）。
- ⚠️1.25：Go 1.8 起混合写屏障（删除+插入）消除了栈重扫，STW 亚毫秒级；1.19+ `GOMEMLIMIT` 防 OOM；**Go 1.25 实验性 Green Tea 分代 GC**（`GOEXPERIMENT=greenteagc`）；`GOGC=off` 可关停。

### 11.4 性能优化心法

- 💡 优化前先测量〔MG〕：pprof CPU/heap profile 定位热点，`go tool trace` 看调度；「揪出隐藏代码」——用 profiler 反查某功能来自哪个包。
- 高频手段：切片/map **预分配**；循环内字符串拼接用 `strings.Builder`〔MG〕；range 大切片免 value 拷贝〔GEP〕；`sync.Pool` 复用临时对象（呼应 EG 的 leaky buffer）〔EG〕；`sort.Slice` → ⚠️1.25 首选 `slices.Sort/SortFunc`（泛型、无反射、更快）〔MG〕。
- 基准测试是性能论据的唯一来源（见 §13）；⚠️ 警惕编译器优化消除「无副作用」的调用（用包级变量接住结果）〔GEP〕。

---

## 12. 反射与 unsafe

### 12.1 反射三定律〔GEP〕

1. 反射可将 interface 变量转换为反射对象：`reflect.TypeOf(x)` / `reflect.ValueOf(x)`；
2. 反射对象可还原成 interface 对象：`v.Interface().(T)`（双括号缺一不可）；
3. 反射对象可修改，**前提是值可设置（addressable）**——必须从指针构建并用 `Elem()`：`reflect.ValueOf(&x).Elem().SetFloat(7.1)`（直接对值副本 Set 会 panic: unaddressable）。

- 本质：反射是**检查 interface 变量的 (value, type) 对**的机制；`type MyInt int` 与 `int` 静态类型不同不能互赋〔GEP〕。

### 12.2 反射三缺点与正当用例〔MG〕

- **三缺点**：①难读难维护；②慢（装箱 + 安全检查开销）；③编译期查不出错——运行时才 panic。**能不用就不用；能类型断言就不反射**〔MG〕。
- 正当用例：处理编写时未知的类型（fmt、template 的实现基础）；**tag 驱动的数据映射**（json Marshal/Unmarshal、ORM——本仓库 model 层 `db/json/pk/table` tag 解析即正当用武之地）〔MG〕。
- ⚠️1.25：Go 1.18+ 泛型让「通用容器/算法」大多不再需要反射；`reflect.TypeFor[T]()`（1.22）可获取泛型类型；结构体比较用 `reflect.DeepEqual`。

### 12.3 unsafe〔MG〕

- `unsafe.Pointer` 可绕过类型系统，**越界访问不报错只返回随机数**——没有安全网。「不确定要不要用就不用」。
- 典型合理场景：cgo 内存互操作、性能敏感零拷贝转换、结构体偏移/对齐计算〔MG〕。
- ⚠️1.25：unsafe.Pointer 有官方「合法使用模式」白名单（Go 1.4+）；把 `uintptr` 存变量再转回指针**不在白名单内**（GC 期间可能悬空）；新代码优先 `unsafe.Add/unsafe.Slice`（1.17）、`unsafe.String/unsafe.StringData/unsafe.SliceData`（1.20）〔MG〕。
- cgo（C 中调用 Go）：`package main` + `import "C"` + `//export FuncName` + `go build -buildmode=c-shared`；引入 cgo 会牺牲「纯 Go 静态编译、随便搬运」的优势〔MG〕。

---

## 13. 测试、基准与工具链

### 13.1 测试类型与命名〔GEP〕

- 文件以 `_test.go` 结尾；函数：`TestXxx(t *testing.T)` / `BenchmarkXxx(b *testing.B)` / `ExampleXxx()`；**测试包建议独立**（`原包名_test`），强制从外部视角使用被测包〔GEP〕。
- 表驱动 + 子测试：`tests := []struct{...}{...}` + `for _, tt := range tests { t.Run(tt.name, ...) }`；子测试命名 `父名/子名`，`-run 父/子` 过滤（**包含匹配**非严格正则，注意误命中）〔GEP〕。
- `t.Parallel()` 并行子测试：实际延迟到**父测试函数体返回后**才开始（barrier/signal 双通道机制），默认并发上限 GOMAXPROCS（`-parallel n` 调整）；并行子测试 + 共享 teardown 的坑：再包一层 `t.Run("group", ...)` 〔GEP〕。
- 方法族谱〔GEP〕：`Error/Errorf`（标记失败继续）/ `Fatal/Fatalf`（失败即退）/ `Skip*`（跳过）；**FailNow/Fatal 只能在当前测试协程内调用**（靠 runtime.Goexit 退出协程）；子测试失败连带父测试失败（Fail 递归上行）。
- `TestMain(m *testing.M)`：声明后由它调度——`retCode := m.Run(); os.Exit(retCode)`；参数需先 `flag.Parse()`；不调 `m.Run()` 则什么都不会跑〔GEP〕。
- 示例测试：`// Output:` 后逐行期望输出；无 Output 则只编译当文档；乱序输出用 `// Unordered output:`；原理是捕获 stdout 再比对〔GEP〕。
- ⚠️1.25：Go 1.18+ 原生模糊测试 `FuzzXxx(f *testing.F)`；`t.Setenv`（1.17）、`t.Cleanup`（1.14）、`t.TempDir`（1.15）；基准日志 1.14 起也需 `-v`。

### 13.2 基准测试〔GEP〕

- 骨架 `for i := 0; i < b.N; i++ { ... }`（⚠️1.25：Go 1.24+ 推荐 `for b.Loop()`——自动防优化删除、循环前代码不计入计时）。
- **b.N 自适应**：先跑 1 次探测，按 ~+20% 递增至跑满 benchTime（默认 1s，`-benchtime` 可改，支持 `Nx` 固定次数），roundUp 到 10 的幂，上限 1e9，结果以最后一次为准〔GEP〕。
- 计时三件套：`b.ResetTimer()`（剔除初始化耗时）；`b.StartTimer()/b.StopTimer()`（分段统计，**累加**）；`b.SetBytes(n)`（报告 MB/s）；`b.ReportAllocs()`（单函数级 -benchmem）〔GEP〕。
- 内存统计 = `ReadMemStats` 净增值 / N（`B/op`、`allocs/op`）〔GEP〕。
- `go test` 结果缓存：`ok xxx (cached)`；**`-count=1` 禁用缓存**（基准永不缓存）〔GEP〕。
- 常用参数：`-bench=.` `-benchmem` `-run` `-v` `-race` `-count` `-parallel` `-timeout`（默认 10m）`-cpu 1,2,4` `-cover -coverprofile` `-failfast` `-json` `-shuffle`（1.17）〔GEP〕。

### 13.3 httptest〔GEP〕〔MG〕

- `httptest.NewRequest` + `httptest.NewRecorder()`：**不起网络直接驱动 handler**，断言 `rec.Code` 与 `rec.Body.String()`——handler 单测首选〔MG〕。
- `httptest.NewServer(handler)`：本地随机端口起真实服务（`defer ts.Close()`），测客户端/集成链路；`NewTLSServer` 自带测试证书〔GEP〕。

### 13.4 pprof 与工具链〔MG〕

- 常驻 HTTP 服务：空导入 `_ "net/http/pprof"` 自动挂 `/debug/pprof/`；自定义 mux 须手动注册 `pprof.Index/Cmdline/Profile/Symbol/Trace`。端点族：`goroutine` / `heap` / `mutex` / `block` / `profile`(CPU) / `trace?seconds=5`〔MG〕。
- 非 web 程序用 `runtime/pprof` 落盘；分析入口 `go tool pprof`（`-http=:8080` 直接开 Web 界面）；`go tool trace` 看执行追踪〔MG〕。
- 💡 用 `ab(1)` 等制造流量后再采样，冷启动数据噪声大〔MG〕。
- `go tool compile -S` / `go build -gcflags -S` 看汇编；`-race` 检测竞态；交叉编译 `GOOS/GOARCH` + `CGO_ENABLED=0`〔MG〕。
- ⚠️1.25：godoc HTTP 服务已非官方重心（pkg.go.dev 为准）；示例函数仍是最有效的可执行文档。

---

## 14. 文件与系统编程

### 14.1 io.Reader / io.Writer 组合惯用法〔MG〕

- 两接口各一个方法却极其强大；实现者遍布标准库（os.File、strings/bytes.Reader、bufio、HTTP body）。
- 💡 **函数签名收 `io.Reader`/`io.Writer` 而非具体类型**——调用方可传文件、内存、网络流（书内把参数写成 `*os.File` 属反例）。
- **包装/装饰**：`bufio.NewReader(f)` 包住 File；`gob.NewEncoder(w)` 包住 Writer——层层叠加互不感知。
- Read 契约：返回 `(n, err)`；**n 可能小于请求长度（短读合法）**，必须用 `buffer[0:n]` 截取实际数据；`io.EOF` 表示结束（⚠️1.25：比较用 `errors.Is(err, io.EOF)` 容忍包装）〔MG〕。

### 14.2 bufio：Scanner vs Reader〔MG〕

- 逐行/逐词/逐字符现代首选 **`bufio.Scanner`**：`scanner.Scan()`/`scanner.Text()`，`ScanLines/ScanWords/ScanRunes` 切分函数。
- ⚠️ Scanner 三坑：**必须检查 `scanner.Err()`**（否则静默丢错）；默认单 token 上限 **64KB**（超长行必须 `scanner.Buffer(buf, max)` 扩容，否则 ErrTooLong）；EOF 处理被隐藏（这同时也是优点）。
- 需精确控制分隔符/超长行/二进制时用 `bufio.Reader`（`ReadString('\n')`/`ReadBytes`）；⚠️ Reader 的 EOF 分支会丢掉没有结尾换行符的最后一行——先处理数据再判 EOF。

### 14.3 文件操作要点〔MG〕

- 写文件五法：`fmt.Fprintf` / `f.WriteString` / **`bufio.Writer` + 必须 `Flush()`**（缓冲写不 Flush 数据滞留，全书最关键坑之一）/ `os.WriteFile`（⚠️1.25：ioutil 已废弃）/ `io.WriteString`。
- `fmt.Fprintf(f, data)` 把数据当格式串（含 `%` 会出事），应 `Fprintf(f, "%s", data)`。
- ⚠️ `defer f.Close()` 应放在 **err 检查之后**（err 非 nil 时 f 为 nil）。
- 按块复制的标准姿势：`buffer := make([]byte, size); n, err := f.Read(buffer); return buffer[0:n]`；流复制直接 `io.Copy(dst, src)`。
- 整读文件 `os.ReadFile` / `io.ReadAll`（替代 ioutil）。

### 14.4 目录遍历与信号〔MG〕

- ⚠️1.25：遍历优先 `filepath.WalkDir`（`fs.DirEntry` 免去每项 Lstat；`filepath.SkipDir` 跳过目录）；单层列目录用 `os.ReadDir`（已排序）。老书 `filepath.Walk` + 手动 Stat 是双重浪费。
- 信号处理标准模式：`sigs := make(chan os.Signal, 1)` + `signal.Notify(sigs)`（不列信号 = 捕获所有，switch 里挑感兴趣的响应）；**SIGKILL/SIGSTOP 不可捕获**。
- ⚠️1.25：首选 `signal.NotifyContext(ctx, sigs...)`（1.16+）+ `srv.Shutdown(ctx)` 做优雅退出；channel 满时信号被丢弃（容量 1 够用）；catch-all 会把运行时内部信号（如 1.14+ 用于抢占的 SIGURG）也接进来——default 分支打印它不是 bug。
- `os.Getuid()` / `user.Current().GroupIds()`；⚠️ syscall 包已冻结，新代码用 `golang.org/x/sys/unix`〔MG〕。

### 14.5 flag 与 CLI〔MG〕

- 基本流程：定义选项 → `flag.Parse()` → 解引用取值；布尔选项必须 `-k=false` 形式（`-k false` 的 false 沦为位置参数）。
- `flag.Var(&v, name, usage)` 接入任意类型（实现 `flag.Value` 接口：`String()/Set(string) error`）。
- ⚠️1.25：现代 CLI 用子命令（flag.NewFlagSet 或 cobra——本仓库 cli 层的 cobra+viper 绑定即此思想的工程化）。

---

## 15. 网络与 HTTP

### 15.1 抽象层级〔MG〕

- 自上而下：`http.Client` → `http.Transport`（`RoundTripper`，含连接池）→ `net.Dial/Listen`（`net.Conn` **同时实现 io.Reader/Writer**）→ raw socket。**选层原则：能用高层就用高层，需要控制力再下沉**。

### 15.2 http.Client / Transport〔MG〕

- 💡 **不要用零值/默认 `http.Client`（无超时）**：`&http.Client{Timeout: 15 * time.Second}` 是最简正确答案，覆盖整个请求生命周期（拨号/TLS/重定向/读 body）。
- 💡 **`http.Client` 并发安全，进程级单例复用，切忌每请求新建**。
- `Transport` 连接池：⚠️ 自定义 Transport 不设 `MaxIdleConnsPerHost` 时**默认仅 2**——高并发同域名频繁重建连接；应 `http.DefaultTransport.(*http.Transport).Clone()` 后微调。`DefaultTransport` 常用值：MaxIdleConns 100 / IdleConnTimeout 90s / TLSHandshakeTimeout 10s。
- ⚠️ `Transport.Dial` 已废弃 → `DialContext: (&net.Dialer{Timeout: t}).DialContext`；请求级取消用 `http.NewRequestWithContext`（1.13+）。
- `httptrace.WithClientTrace` 观察连接池命中（`GotConnInfo.Reused/IdleTime`）〔MG〕。

### 15.3 http.Server：超时与优雅关闭〔MG〕

- **四层超时体系**：`Client.Timeout`（客户端全程）> Transport 各段（Dial/TLS 握手/响应头/Expect-Continue）> `net.Conn.SetDeadline`（连接读写绝对截止，⚠️ 绝对时间过期后所有读写立即失败，每次操作各限时须每次刷新）> **Server `ReadTimeout/WriteTimeout`**（防慢客户端）。
- ⚠️1.25：服务端生产建议同时显式设置 `ReadHeaderTimeout`（专防 slowloris）与 `IdleTimeout`。
- ⚠️ `http.ListenAndServe` 短路写法**无法优雅关闭**：保留 `srv` 实例，`signal.NotifyContext` + `srv.Shutdown(ctx)` 排空在途请求后退出；`ListenAndServe` 返回 `http.ErrServerClosed` 属正常路径。
- 模板：`template.Must(template.ParseGlob(...))`（失败即 panic，适合启动期校验）；正式代码一律 `html/template`（自动转义防 XSS），不要 fmt 拼 HTML。
- ⚠️1.25：`net/rpc` 已被官方冻结（仅 gob、无超时/重试/流式）——新项目选 gRPC 或 HTTP/JSON；`rand.Seed` 已废弃（全局源自动播种，安全场景用 crypto/rand）。
- accept-goroutine 模式的收尾〔MG〕：`close(l)` 使 Accept 返回 `ErrClosed`（用 `errors.Is` 判断），WaitGroup 等在途 goroutine。

---

## 16. 定时器与资源泄露〔GEP〕

- Timer（一次性）vs Ticker（周期性）：runtime 层唯一差别是 `period` 是否为 0；`sendTime` 用 `select+default` 保证**永不阻塞**——消费方跟不上时**静默丢弃 tick**（drop ticks on the floor），不堆积不补偿。
- `Stop()` 只摘除定时器**不关闭管道 C**；Timer 的 `Reset` 官方约定：只应作用于「已过期或已 Stop」的 Timer，**返回值不可靠不应依赖**；select 提前返回务必 `defer timer.Stop()`。
- 💡 铁律：**`ticker := time.NewTicker(d)` 之后立刻 `defer ticker.Stop()`**——即使你的函数永不退出不泄露，**别人拷走这段代码放进会退出的函数里就泄露了**。
- `NewTicker(d<=0)` 直接 panic；`time.Tick(d)` 无法 Stop，仅适合「永不停止」的轮询；⚠️ 经典错误：`for { select { case <-time.Tick(1s) } }`——每轮 select 都**新建**一个 Ticker，越积越多。
- **go-fastping 资源泄露案例**（管理 1000 台服务器 4 天后 CPU 100%）：pprof CPU profile 见 `runtime.siftdownTimer` 热点 → 调用栈定位开源库 `Run()` 的多个出口都没 Stop 两个 Ticker → 每次探测泄露 2 个定时器 → 累积数百万僵尸定时器耗尽 CPU。修复 = `defer timeout.Stop()`。**诊断方法论**：CPU 持续缓慢升高 → pprof top → timer 类热点函数 → 调用栈归属 → 找未 Stop 的定时器。
- ⚠️1.25：定时器架构已重写两次（Go 1.14 起每 P 一个定时器堆，无独立 timerproc；1.23 再重构）；**Go 1.23 起 Timer/Ticker 的 C 改为无缓冲，且不再被引用的定时器可被 GC（即使未 Stop）**、「Reset 后收到两个事件」的旧风险被根治——但显式 Stop 仍是最佳实践。

---

## 17. 语法糖：`:=` 与 `...`〔GEP〕

### 17.1 简短变量声明 `:=`

- **只能用于函数内部**（等价「声明+赋值」，而赋值语句不允许出现在函数外）。
- 左侧必须**至少一个新变量**（否则编译错误 `no new variables on left side of :=`）；同作用域内旧变量被「重新声明」（实为重新赋值，无害）。
- ⚠️ **跨作用域同名即新变量（遮蔽）——缺陷高发点**：if/for 块内 `field, err := ...` 的 err 是新变量，外层读不到内层赋值。最佳实践：块内重复 `:=` 同名变量（尤其 err）时停下来确认是否真想遮蔽；要在内层修改外层变量用 `=`。
- 借助 `go vet` 的 shadow 检查辅助发现遮蔽。

### 17.2 可变参 `...`

（见 §4.2）核心坑：`s...` 展开与函数内部切片**共享底层数组**——把可变参切片当只读入参，要修改先克隆。

---

## 18. Go 1.9/1.12 → Go 1.25 演进速查

| 主题 | 旧书说法（1.9/1.12） | Go 1.25 现状 |
|---|---|---|
| for 循环变量 | 单一变量，闭包捕获共享 | **1.22 起每轮迭代独立作用域**；新增 range over int / range-over-func |
| 切片扩容 | <1024 翻倍，≥1024 ×1.25 | 1.18 起渐进公式（小容量翻倍、大容量 ≈1.25 平滑增长，size class 对齐） |
| map 底层 | bmap 桶 + overflow 链 | **1.24 起 Swiss Table**；负载因子 6.5、非并发安全结论不变 |
| defer | 链表实现、昂贵 | 1.14 起 open-coded，近乎零开销 |
| GMP | 协作式抢占 | **1.14 起基于信号（SIGURG）异步抢占**；sysmon；netpoller 管网络阻塞 |
| GOMAXPROCS | 默认 = CPU 核数 | **容器感知**（cgroup CPU 配额） |
| GC | 三色 + 写屏障 | 混合写屏障（1.8）、GOMEMLIMIT（1.19）、实验性 Green Tea 分代 GC（1.25） |
| unsafe | uintptr 自由运算 | 官方合法模式白名单；优先 unsafe.Add/Slice/String（1.17/1.20） |
| 排序 | sort.Slice | 首选泛型 `slices.Sort/SortFunc` + `cmp`（1.21+） |
| strings | Title、Replace(-1) | Title 已弃用；ReplaceAll、Cut（1.18）、Builder（1.10）、Clone、Lines（1.24） |
| 随机数 | rand.Seed(time.Now().Unix()) | **rand.Seed 已废弃**（全局源自动播种）；math/rand/v2（1.22）；安全场景 crypto/rand |
| ioutil | ioutil.ReadFile/WriteFile | os.ReadFile / os.WriteFile / io.ReadAll / io.Discard |
| 目录遍历 | filepath.Walk + Stat | **filepath.WalkDir** + fs.DirEntry + SkipDir |
| 信号 | Notify + channel | 首选 **signal.NotifyContext** + Shutdown 优雅退出 |
| HTTP 取消 | Transport.CancelRequest | **http.NewRequestWithContext**；Client.Timeout |
| 定时器 | 64 桶 + timerproc；未 Stop 永久泄露 | 每 P 定时器堆；**1.23 起未引用可被 GC、C 无缓冲** |
| 测试 | 单测/基准/示例 | + 模糊测试（1.18）、`b.Loop()`（1.24）、t.Setenv（1.17）、synctest（1.25 实验） |
| 容器/泛型 | interface{} + 手写容器 | **泛型（1.18）** + slices/maps/cmp 标准库 |
| interface{} | interface{} | `any` 别名 |
| GOPATH 安装 | ~/go/src + go install | Go modules + GOTOOLCHAIN 自动切换 |
| net/rpc | rpc.Register/ServeConn | **包已冻结**；新项目 gRPC / HTTP+JSON |

---

## 19. 本项目（go-ai-scaffold）对照审计清单

> 以下条目把上述知识图谱映射到本仓库技术栈（gin + viper + cobra + sqlx/squirrel + PostgreSQL，Go 1.25），作为代码审计的检查基线。项目自身约定（AGENTS.md）优先级最高，此处仅列「约定之外仍值得核对」的知识点。

**架构与分层**

- [ ] `pkg/*` 不 import `mod/*`（接口定义在消费方、pkg 可独立复用的落地）。
- [ ] controller 只做 `BindForm → service → JsonSuccess`，不直连 dao（分层=接口边界，类比 io.Reader/Writer 收窄）。
- [ ] service 不触碰 `*context.Context`（HTTP 语义的 ctx 只留在 controller/middleware 层）。

**数据与 SQL**

- [ ] dao 一律参数化（`Where("col=?", v)`），无 `fmt.Sprintf` 拼 SQL（参数化=安全底线）。
- [ ] 大列表查询注意 slice/map 预分配；遍历结果转换时避免循环内 `+=` 拼字符串（用 strings.Builder）。
- [ ] model tag 语法逐字段核对（`key:"value"` 冒号前后无空格——写错是静默失败）。

**并发与资源**

- [ ] 所有 `go func` 启动点：是否有退出路径（context/channel/WaitGroup），防 goroutine 泄漏。
- [ ] 所有 `time.NewTicker/NewTimer`：创建后是否紧跟 `defer Stop()`。
- [ ] 所有锁：`Lock()` 后是否 `defer Unlock()`；RWMutex 只用于读多写少且读临界区较重的路径。
- [ ] mqtt/redis/aikit 等长连接 kit：`select` 循环是否有 `ctx.Done()` 分支；nil channel 禁用分支、close 广播等惯用法是否正确使用。

**错误处理**

- [ ] panic(exception.New(...)) 仅用于业务错误、由 `middleware.Recover` 统一兜底（对应「包内 panic 转 error」模式）；**不得在 goroutine 内 panic 而无人 recover**。
- [ ] 无 `_ = err` 吞错误；`resp.Body` 类资源有 Close。
- [ ] 删除操作置空唯一字段（phone/username）防脏数据。

**安全**

- [ ] `Pwd` 类敏感字段 `json:"-"`；新代码 bcrypt/scrypt/argon2（存量 MD5 是已知债务）。
- [ ] 无硬编码密钥/连接串。

**HTTP 服务**

- [ ] gin/http.Server 是否配置 ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout；是否支持优雅关闭（Shutdown）。
- [ ] 出站 http.Client 是否复用单例并设置 Timeout；自定义 Transport 的 MaxIdleConnsPerHost 是否 >2。

**工程化**

- [ ] `go build ./... && go vet ./... && gofmt -l .` 全绿（AGENTS.md 规定的验证基线）。
- [ ] 是否存在测试（当前仓库无测试——表驱动 + httptest.NewRecorder 是补测起点）；CI 是否含 `go test -race`。

---

*本文档由三份资料精读融合生成：Effective Go（官方在线版全文精读）+《Go专家编程》全 11 章笔记 +《Mastering Go》中文第一版全 13 章笔记（该 PDF 为删节版，第 11 章正文缺失、已据目录与交叉引用重建）。生成于 2026-08-22。*
