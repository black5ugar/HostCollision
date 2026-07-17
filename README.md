# HostCollision

[🇨🇳 中文版点此](#中文版说明)

HostCollision is a high-performance tool for discovering virtual hosts by sending HTTP requests with customized `Host` headers to the same IP address.  
It is commonly used in penetration testing to detect websites behind reverse proxies, shared hosting, CDN environments, wildcard DNS, and misconfigured vhosts.

This version is a fully refactored and modularized implementation based on the original project.



## ✨ Features

- 🚀 **High-speed scanning** powered by goroutine worker pool  
- 🧠 **Dedicated baseline probes** to filter out generic/wildcard/default pages without consuming a real candidate
- 📝 **Real-time terminal logs** showing status code, duration, and similarity  
- 📄 **CSV result output** for better data analysis  
- ⚙️ **Configurable parameters** (threads, sleep, threshold, max hits per IP…)  
- 🛡️ **Safer HTTP handling** with redirect blocking, redirect-aware fingerprints, cancellation support, and a 10 MiB response limit
- 🎨 **Beautiful CLI banner**



## 📦 Installation

Go 1.18 or later is required.

```bash
go build -o hostcollision ./cmd/hostcollision
```

Run the tests with:

```bash
go test ./...
```



## 🧪 Example Usage

```
./hostcollision \
  -i ip.txt \
  -d host.txt \
  -o output.csv \
  -n 20 \
  -s 0 \
  -m 10 \
  -r 85
```



## 🗂 Command-line Options


| Option | Description                                 |
| ------ | ------------------------------------------- |
| `-i`   | Path to IP list file (required)             |
| `-d`   | Path to host dictionary file (required)     |
| `-o`   | Output CSV file path (required)             |
| `-n`   | Number of goroutines (default 20)           |
| `-s`   | Sleep between requests in ms (default 1000) |
| `-m`   | Max retained hosts per IP (default 50)      |
| `-r`   | Similarity threshold (0–100, default 85)    |



## 📤 Output

output.csv will contain:
```
ip,host,status,length,similar
127.0.0.1,www.example.com,200,7648,32
```

`ip`      – target IP address

`host`    – tested Host header

`status`  – HTTP status code

`length`  – response body length (bytes)

`similar` – similarity score (0–100) compared to baseline for that IP

Only responses with status codes from 200 through 399 and similarity below the configured threshold are retained. Before scanning the dictionary, HostCollision sends `Host: hostcollision-baseline.invalid` to each IP and uses that response as the default-page baseline. Baseline probes run with the configured concurrency limit. Response bodies are used for the normal similarity score; for redirects, the status code and `Location` header are also compared so that different empty-body redirects are not incorrectly treated as identical.

For safety, redirects are not followed, response bodies larger than 10 MiB are rejected, and `Ctrl+C` cancels outstanding requests. If a baseline probe fails, candidates for that IP are still scanned with a similarity score of `0`.



## 🧪 Minimal Local Test Setup

This is a simple way to verify the tool works end-to-end on your machine.

1. **Create `ip.txt`**

   ```
   127.0.0.1
   ```

2. **Create** `host.txt`
	
	```
	www.aaa.com
	www.bbb.com
	www.ccc.com
	```

3. **Run a simple HTTP server** (example)

   The repository includes a sample server in `testserver/`. It listens on port 80, which may require elevated privileges:

   ```bash
   go run ./testserver
   ```

   You can also use any HTTP server that returns different content based on the `Host` header. A target entry may include a port, for example `127.0.0.1:8080`. Bare IPv6 addresses such as `2001:db8::1` are supported; use brackets when specifying an IPv6 port, for example `[2001:db8::1]:8080`.

4. **Start scanning**

   ```
   ./hostcollision -i ip.txt -d host.txt -o output.csv -n 3 -s 0 -m 10 -r 85
   ```

​	You will see real-time logs in the terminal and structured results in `output.csv`.



## 📚 Project Structure

```
cmd/hostcollision        # Main entry point (CLI)
internal/app             # Application orchestration (read -> scan -> write)
internal/config          # CLI configuration parsing and validation
internal/scanner         # Core scanning logic (workers, HTTP, thresholds)
internal/similarity      # Similarity engine for response body comparison
internal/iohelper        # File reading/writing utilities
internal/banner          # CLI banner (version, author, GitHub)

```



## ⚠️ Legal & Ethical Disclaimer

This tool is intended **only for authorized security testing and research**.
 Do **not** use it against targets without explicit permission.
 You are solely responsible for complying with all applicable laws and regulations.



# 中文版说明

[🇬🇧 English version click here](#HostCollision)

HostCollision 是一个通过自定义 `Host` 头，对目标 IP 进行批量请求，从而发现隐藏虚拟主机的高性能扫描工具。
 常见使用场景包括：

- 反向代理 / 共享主机环境中的站点枚举
- CDN 场景下真实站点的探测
- 泛解析 / 默认站点识别
- Vhost 配置错误排查

当前版本对原项目进行了重构，结构更加清晰、模块化，便于维护和扩展。



## ✨ 功能特点

- 🚀 **高并发扫描**：基于 goroutine 的 worker pool
- 🧠 **响应相似度检测**：过滤统一错误页 / 默认页 / 泛解析内容
- 🎯 **独立基准探测**：使用专门的无效 Host 获取基准，不占用真实字典候选
- 📡 **终端实时日志**：显示 IP、Host、状态码、耗时、相似度、过滤原因
- 📄 **CSV 结果输出**：带表头，方便后续用 Excel / 脚本分析
- ⚙️ **可配置参数**：线程数、请求间隔、相似度阈值、每 IP 最大命中数等
- 🛡️ **安全的 HTTP 处理**：禁止跨目标重定向、识别不同的重定向目标、支持取消，并限制响应体大小





## 📦 安装方式

需要 Go 1.18 或更高版本。

```bash
go build -o hostcollision ./cmd/hostcollision
```

运行测试：

```bash
go test ./...
```



## 🧪 使用示例

```
./hostcollision \
  -i ip.txt \
  -d host.txt \
  -o output.csv \
  -n 20 \
  -s 1000 \
  -m 50 \
  -r 85

```





## 🗂 参数说明

| 参数 | 说明                                                         |
| ---- | ------------------------------------------------------------ |
| `-i` | IP 列表文件路径（必选）                                      |
| `-d` | Host 字典文件路径（必选）                                    |
| `-o` | 输出 CSV 文件路径（必选）                                    |
| `-n` | 并发 goroutine 数量（默认 `20`）                             |
| `-s` | 每次请求间的 sleep（毫秒，默认 `1000`）                      |
| `-m` | 单个 IP 最多保留的 Host 数（默认 `50`）                      |
| `-r` | 相似度阈值（0–100，默认 `85`，大于等于该值认为“过于相似”而被过滤） |



## 📤 输出说明

结果文件为 CSV 格式，包含表头：

```
ip,host,status,length,similar
127.0.0.1,www.example.com,200,7648,32
```

字段含义：

- `ip`      – 被扫描的 IP
- `host`    – 请求使用的 Host 头
- `status`  – HTTP 状态码
- `length`  – 响应 Body 长度（字节）
- `similar` – 与该 IP 基准响应的相似度（0–100）

程序仅保留状态码为 200–399 且相似度低于指定阈值的响应。扫描字典前，HostCollision 会向每个 IP 发送一次 `Host: hostcollision-baseline.invalid` 请求，以其响应作为默认页面基准；基准请求同样受并发数限制。普通响应使用 Body 计算相似度；对于重定向响应，还会比较状态码和 `Location` 响应头，避免把 Body 为空但跳转目标不同的重定向错误地过滤掉。

为避免请求离开扫描范围，程序不会跟随重定向；超过 10 MiB 的响应体会被拒绝，按 `Ctrl+C` 可取消尚未完成的请求。如果某个 IP 的基准请求失败，该 IP 仍会继续扫描，候选响应的相似度记为 `0`。



## 🧪 最小本地测试环境

1. **准备 `ip.txt`**

```
127.0.0.1
```

2. **准备 `host.txt`**

```
www.aaa.com
www.bbb.com
www.ccc.com
```

3. **启动本地 HTTP 服务**（参见 `testserver/`）

示例服务监听 80 端口，可能需要额外权限：

```bash
go run ./testserver
```

也可以使用其他根据 `Host` 头返回不同页面的 HTTP 服务。IP 列表支持携带端口，例如 `127.0.0.1:8080`。同时支持裸 IPv6 地址，例如 `2001:db8::1`；IPv6 携带端口时请使用方括号，例如 `[2001:db8::1]:8080`。

4. **执行扫描**

```
./hostcollision -i ip.txt -d host.txt -o output.csv -n 3 -s 0 -m 10 -r 85
```

终端可以看到实时日志，结果会以 CSV 形式写入 `output.csv`。



## 📚 项目结构

```
cmd/hostcollision        # 程序入口（main）
internal/app             # 扫描流程编排：读入 -> 扫描 -> 写结果
internal/config          # 命令行参数解析与配置校验
internal/scanner         # 核心扫描逻辑（并发、HTTP、阈值控制）
internal/similarity      # 相似度计算模块
internal/iohelper        # 文件读写工具
internal/banner          # 终端 Banner 展示
```



## ⚠️ 法律与合规声明

本工具仅供 **授权的安全测试与研究使用**。
请勿在未获得明确授权的前提下，对任何目标使用本工具。
使用本工具产生的一切后果由使用者自行承担。
