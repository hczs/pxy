# px

[English](./README.en.md)

[![CI](https://github.com/hczs/px/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/px/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/hczs/px)](https://github.com/hczs/px/releases)
[![License](https://img.shields.io/github/license/hczs/px)](./LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/hczs/px.svg)](https://pkg.go.dev/github.com/hczs/px)

`px` 是一个轻量的 Go CLI，用来在当前终端会话中快速开启、关闭代理环境变量。

它会自动识别常见本地代理工具，保存本地配置，并安装 shell function，让 `px on` / `px off` 可以真正影响当前 shell。

## 特性

- 一条命令初始化：自动检测代理配置并写入 shell profile。
- 当前会话生效：通过 shell function 修改当前终端环境变量。
- 可恢复现场：`px off` 会恢复 `px on` 之前已有的代理变量。
- 支持自动检测：Clash、Clash Verge、Surge、v2rayA、v2rayN。
- 支持多种 shell：bash、zsh、PowerShell。
- 小而透明：配置保存在 `~/.px/config.yaml`，命令输出直接可读。

## 安装

从 [GitHub Releases](https://github.com/hczs/px/releases) 下载对应平台的二进制文件。

也可以从源码安装：

```bash
go install github.com/hczs/px@latest
```

Release 构建支持 macOS、Linux、Windows 的 amd64 / arm64。

## 快速开始

初始化：

```bash
px init
```

按提示选择自动检测到的代理配置，或手动输入代理地址。完成后重启终端，或执行命令输出中的 `source ...`。

开启代理：

```bash
px on
```

查看状态：

```bash
px status
```

测试当前出口 IP：

```bash
px test
```

关闭代理：

```bash
px off
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `px init` | 检测 shell 和代理配置，保存配置并安装 shell function |
| `px on` | 在当前 shell 中开启代理环境变量 |
| `px off` | 恢复或清理当前 shell 中的代理环境变量 |
| `px status` | 查看当前代理环境变量状态 |
| `px test` | 通过 `https://ipinfo.io/json` 测试当前代理 |
| `px list` | 列出检测到的本地代理软件 |
| `px config` | 手动重新配置代理 |

## 支持的环境变量

`px on` 会根据配置设置以下变量：

- `http_proxy` / `HTTP_PROXY`
- `https_proxy` / `HTTPS_PROXY`
- `all_proxy` / `ALL_PROXY`

HTTP 与 HTTPS 代理使用 `http://host:port`，SOCKS5 代理使用 `socks5://host:port`。

## 工作方式

普通 CLI 子进程无法直接修改父 shell 的环境变量，所以 `px init` 会向 shell profile 写入一个 `px` shell function。

之后执行 `px on` / `px off` 时，shell function 会调用 `px` 生成环境变量脚本，并在当前 shell 中执行。

## 配置

默认配置文件：

```text
~/.px/config.yaml
```

示例：

```yaml
proxy:
  http:
    enabled: true
    host: 127.0.0.1
    port: 7890
  https:
    enabled: true
    host: 127.0.0.1
    port: 7890
  socks5:
    enabled: true
    host: 127.0.0.1
    port: 7890
  auto_detect: true
  source: 自动检测(Clash)
```

## 开发

运行本地检查：

```bash
make check
```

等价命令：

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

## 发布

创建并推送版本标签：

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions 会运行测试，并通过 GoReleaser 发布构建产物。

## 贡献

欢迎提交 Issue 和 Pull Request。提交前请运行：

```bash
make check
```

如果修改 CLI 行为，请同步更新 README、help 文案和相关测试。

## 许可证

[MIT](./LICENSE)
