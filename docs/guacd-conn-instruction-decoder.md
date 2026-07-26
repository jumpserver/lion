# Guacd 连接指令解析与读取改进记录

## 基本信息

- 日期：2026-07-26
- 基线分支：`dev`
- 基线提交：`21da7bc`
- 工作分支：`feat_guacd_stream_decoder`
- 范围：Go 侧 guacd TCP 连接、通用 Guacamole 指令解析、回放文件指令读取

## 原实现与问题

连接读取原来的处理链路如下：

```text
net.Conn
  -> bufio.Reader.ReadString(';')
  -> 拼接到累计字符串
  -> ParseInstructionString(累计字符串)
  -> 解析成功后返回一条 Instruction
```

Guacamole 指令并不是简单的分号分隔文本，而是长度前缀协议：

```text
LENGTH.VALUE[,LENGTH.VALUE...];
```

`LENGTH` 表示 `VALUE` 的 Unicode 码点数量。逗号和分号都允许出现在
`VALUE` 内部，因此只有读取完长度指定的内容之后，后续的逗号或分号才是
结构分隔符。

原实现存在以下问题：

1. `ReadString(';')` 不理解长度前缀，会把参数内的分号当成候选边界。例如
   `4.echo,3.a;b;` 会先读到 `4.echo,3.a;`，解析失败后再继续读取。
2. 每遇到参数内分号就会重新解析完整累计字符串；`ret += msg` 也会反复复制。
   对包含多个分号或较长 blob 的指令，时间复杂度和分配次数都会增加。
3. 累计字符串没有上限。畸形或恶意长度前缀可以让读取端持续等待和增长内存。
4. 每次候选解析失败都会把累计协议内容和错误写入日志。合法的参数内分号也会
   产生错误日志，blob、剪贴板等内容还可能形成大日志或暴露不应记录的数据。
5. 15 秒读取截止时间在每个 `ReadString(';')` 前刷新，单条指令可通过持续发送
   含分号的碎片长期延长读取时间。
6. `ParseInstructionString()` 会把 `3.one;3.two;` 这类多条指令错误解释成一条
   opcode 为 `one`、参数为 `two` 的指令，没有拒绝尾随数据。
7. 指令序列化时，参数长度按 Unicode 码点计算，但 opcode 长度按 UTF-8 字节数
   计算，非 ASCII opcode 无法被符合协议的解析器正确读取。
8. 回放文件读取中另有一份相同的“按分号读取、失败后重试”逻辑，行为和风险重复。

## 协议依据与限制

Apache Guacamole 协议说明明确要求长度按 Unicode 字符/码点计算，并指出该格式
应当按流增量解析：

- <https://guacamole.apache.org/doc/gug/guacamole-protocol.html>
- <https://github.com/apache/guacamole-server/blob/main/src/libguac/guacamole/parser-constants.h>
- <https://github.com/apache/guacamole-server/blob/main/src/libguac/guacamole/parser.h>
- <https://github.com/apache/guacamole-server/blob/main/src/libguac/parser.c>

新解码器采用与当前 libguac 解析器一致的边界：

| 限制 | 值 | 目的 |
| --- | ---: | --- |
| 单个元素最大 Unicode 码点数 | 8192 | 拒绝异常长度声明 |
| 长度前缀最大位数 | 5 | 限制长度字段 |
| 单条指令最大元素数（含 opcode） | 128 | 限制参数数量 |
| 单条指令最大 UTF-8 字节数 | 32768 | 对应 8192 个四字节码点的最坏情况缓冲上限 |

## 改进方案

### 1. 共用的增量解码器

新增 `guacd.InstructionDecoder`，直接从 `io.Reader` 按以下状态推进：

1. 只读取 ASCII 数字形式的长度前缀，遇到点号后确定元素长度。
2. 使用 `ReadRune()` 精确读取指定数量的 Unicode 码点，并验证 UTF-8。
3. 内容读完后只接受逗号或分号；逗号进入下一个元素，分号完成当前指令。
4. 全程统计长度位数、元素数量和累计字节数，越界立即返回分类错误。

每个输入字节/码点只参与一次结构解析，读取复杂度为 O(n)，内存使用受协议
上限约束。参数内的逗号、分号不会再触发试探解析或错误日志。

### 2. 明确 EOF 与 I/O 错误语义

- 新指令尚未读取任何字节时遇到流结束，返回 `io.EOF`。
- 指令读取到一半时结束，返回可通过 `errors.Is(err, io.ErrUnexpectedEOF)`
  识别的分类错误。
- 网络超时等底层 I/O 错误原样保留，不误分类为协议内容错误。
- 无效 UTF-8、错误长度、错误终止符、超限和尾随数据分别返回明确错误。

### 3. 连接读取

`Tunnel` 持有并复用一个解码器，`ReadInstruction()` 不再按分号试探边界，也不再
记录预期中的中间解析失败。读取截止时间在每条指令开始时设置一次，因此 15 秒
限制覆盖整条指令，而不是每个分号碎片。

公开的 `ReadInstruction()`、`Read()`、`Tunneler` 接口和正常指令返回结构保持不变。

### 4. 字符串解析与回放读取

- `ParseInstructionString()` 复用相同解码器，并要求输入严格包含一条指令；
  多条指令或其他尾随数据会被拒绝。
- 回放文件读取复用相同解码器，删除重复的按分号累计和错误日志逻辑。
- `Instruction.String()` 对 opcode 和参数都使用 Unicode 码点数生成长度前缀。

## 文件改动记录

| 文件 | 改动 |
| --- | --- |
| `pkg/guacd/instruction_decoder.go` | 新增有界、长度驱动的流式解码器及 EOF/I/O 处理 |
| `pkg/guacd/instruction.go` | 字符串解析改为复用解码器；新增分类错误；修正 Unicode opcode 长度 |
| `pkg/guacd/tunnel.go` | `Tunnel` 复用解码器；每条指令仅设置一次读取截止时间 |
| `pkg/guacd/instruction_test.go` | 重写解析测试，覆盖合法、畸形、边界、碎片流、连接、竞态和 fuzz 场景 |
| `pkg/tunnel/replay_part_upload.go` | 回放读取改为复用通用解码器 |
| `docs/guacd-conn-instruction-decoder.md` | 记录分析、设计、行为变化、文件清单和验证结果 |

## 行为变化与兼容性

- 正常的 ASCII/Unicode Guacamole 指令及公开读取接口保持兼容。
- 参数内包含逗号或分号时，改为一次正确读取，不再试探和打印错误日志。
- 畸形输入不再通过继续拼接后续数据尝试恢复，而是立即返回协议错误；连接调用方
  原本已在读取错误时退出，因此与现有错误处理路径一致。
- `ParseInstructionString()` 现在拒绝多条指令和尾随内容，避免静默误解析。
- 读取超时常量仍为 15 秒，但语义收紧为单条完整指令的总读取时间。

## 验证记录

已执行：

```text
go test ./pkg/guacd -count=1
go test -race ./pkg/guacd -count=1
go test ./pkg/guacd -run '^$' -fuzz=FuzzInstructionRoundTrip -fuzztime=3s
go test ./... -count=1
go vet ./...
go test -race ./... -count=1
```

覆盖的关键场景：

- 每次只返回一个字节的碎片化输入；
- 同一流内连续两条指令；
- 参数内容包含逗号和分号；
- 多字节 Unicode opcode 和参数；
- 空 opcode、空参数；
- 长度、内容和终止符位置发生 EOF；
- 底层 reader 返回非 EOF 错误；
- 无效 UTF-8、非数字/超长长度、元素数和总字节数越界；
- 通过 `net.Pipe` 分片写入的 `Tunnel.ReadInstruction()` 集成路径；
- 单条指令含多个内部分号时只设置一次读取截止时间；
- 编码后再解析的模糊测试。

上述定向测试、全仓测试、静态检查和全仓竞态测试均通过。
