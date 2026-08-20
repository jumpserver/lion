# Guacd 指令解码器

## 协议格式

Guacamole 指令由一个 opcode 和零个或多个参数组成。每个元素均使用
`LENGTH.VALUE` 格式，元素之间以逗号分隔，整条指令以分号结束。

`LENGTH` 表示 `VALUE` 的 Unicode 码点数，不是 UTF-8 字节数。逗号和分号可以
出现在元素内容中，只有读取完声明数量的码点后，后续逗号或分号才作为结构分隔符。

## InstructionDecoder

`InstructionDecoder` 从 `io.Reader` 增量读取指令。`ReadInstruction()` 每次只返回
一条完整的 `Instruction`，并保留缓冲区中属于下一条指令的数据。

创建解码器时：

- 输入已经是 `*bufio.Reader` 时直接复用；
- 其他 `io.Reader` 会由新的 `bufio.Reader` 包装；
- 对非缓冲 reader，应持续复用同一个解码器，避免丢失预读数据；
- 解码器不用于并发读取。

单条指令的读取过程如下：

1. 读取非空的 ASCII 十进制长度前缀；
2. 按声明长度读取 UTF-8 编码的 Unicode 码点；
3. 验证元素后的字符为逗号或分号；
4. 遇到逗号时继续读取下一个元素，遇到分号时返回指令。

## 解析限制

解码器使用以下固定限制：

| 限制 | 值 |
| --- | ---: |
| 单个元素最大 Unicode 码点数 | 8192 |
| 长度前缀最大位数 | 5 |
| 单条指令最大元素数（包含 opcode） | 128 |
| 单条指令最大累计字节数 | 32768 |

累计字节数包含长度前缀、元素内容和分隔符。超过任一限制都会立即终止当前指令的
解析。

## 错误语义

解析错误通过包内定义的错误值分类：

| 错误 | 含义 |
| --- | --- |
| `ErrInstructionMissSemicolon` | 字符串解析输入未以分号结束 |
| `ErrInstructionMissDot` | 长度前缀后缺少点号 |
| `ErrInstructionBadDigit` | 长度为空或包含非数字字符 |
| `ErrInstructionBadContent` | 元素内容不完整 |
| `ErrInstructionBadTerminator` | 元素后的字符不是逗号或分号 |
| `ErrInstructionTooLong` | 长度前缀、元素长度或累计字节数超限 |
| `ErrInstructionTooManyElements` | 指令元素数量超限 |
| `ErrInstructionInvalidUTF8` | 元素包含无效 UTF-8 |
| `ErrInstructionTrailingData` | 单指令字符串后仍有其他数据 |

新指令尚未读取任何字节时遇到流结束，返回 `io.EOF`。长度、内容或终止符读取到
一半时遇到 EOF，返回的错误可通过 `errors.Is(err, io.ErrUnexpectedEOF)` 判断，
同时保留对应的解析错误分类。网络超时等非 EOF 的底层读取错误原样返回。

解析错误发生时，当前指令已经消费的输入不会回退，调用方应结束当前流或连接。

## 字符串解析与序列化

`ParseInstructionString()` 使用 `InstructionDecoder` 解析输入，并要求输入严格包含
一条完整指令：

- 输入必须以分号结束；
- 第一条指令之后必须立即到达 EOF；
- 多条指令或任何尾随内容均返回 `ErrInstructionTrailingData`。

`Instruction.String()` 对 opcode 和所有参数统一按 Unicode 码点数生成长度前缀。
生成结果缓存在 `ProtocolForm` 中，后续调用直接返回缓存内容。

## Tunnel 集成

`Tunnel` 在建立连接时基于其 `bufio.Reader` 创建一个解码器，并在连接生命周期内
持续复用。`Tunnel.ReadInstruction()` 在每条指令开始读取前设置一次 15 秒读取
截止时间，然后由解码器读取并返回一条指令。

`Tunnel.Read()` 复用 `ReadInstruction()`，并将解析后的指令重新序列化为字节。
握手阶段的 `expect()` 同样通过该读取路径校验期望的 opcode。

## 回放文件集成

回放文件读取函数接收现有的 `*bufio.Reader`，并通过 `InstructionDecoder` 逐条读取
指令。由于解码器直接复用该 reader，多次调用不会丢失 reader 中已缓冲的后续数据。
回放时间扫描据此遍历指令并提取 `sync` 时间。
