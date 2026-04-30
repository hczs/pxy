# pxy

[English](./README.en.md)

[![CI](https://github.com/hczs/pxy/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/pxy/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/hczs/pxy)](https://github.com/hczs/pxy/releases)
[![License](https://img.shields.io/github/license/hczs/pxy)](./LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/hczs/pxy.svg)](https://pkg.go.dev/github.com/hczs/pxy)

`pxy` 是一个轻量的 Go CLI，用来在当前终端会话中快速开启、关闭代理环境变量。

它会自动识别常见本地代理工具，保存本地配置，并安装 shell function，让 `pxy on` / `pxy off` 可以真正影响当前 shell。

## 特性

- 一条命令初始化：自动检测代理配置并写入 shell profile。
- 当前会话生效：通过 shell function 修改当前终端环境变量。
- 可恢复现场：`pxy off` 会恢复 `pxy on` 之前已有的代理变量。
- 支持自动检测：Clash、Clash Verge、Surge、v2rayA、v2rayN。
- 支持多种 shell：bash、zsh、PowerShell。
- 小而透明：配置保存在 `~/.pxy/config.yaml`，命令输出直接可读。

## 安装

从 [GitHub Releases](https://github.com/hczs/pxy/releases) 下载对应平台的二进制文件。

不知道下载哪个文件时：

- Apple Silicon Mac 下载 `darwin_arm64`
- Intel Mac 下载 `darwin_amd64`
- 常见 Linux x86_64 下载 `linux_amd64`
- Linux ARM64 下载 `linux_arm64`
- 常见 Windows x86_64 下载 `windows_amd64.zip`
- Windows ARM64 下载 `windows_arm64.zip`

### macOS 二进制安装

以下命令以 Apple Silicon Mac 为例。Intel Mac 把 `darwin_arm64` 换成 `darwin_amd64`。

```bash
cd ~/Downloads
tar -xzf pxy_VERSION_darwin_arm64.tar.gz
chmod +x pxy
sudo mv pxy /usr/local/bin/pxy
pxy init
```

如果系统提示文件来自未知开发者，执行：

```bash
xattr -d com.apple.quarantine /usr/local/bin/pxy
```

然后重启终端，或执行 `pxy init` 输出的重新加载命令。之后即可使用：

```bash
pxy on
pxy status
pxy off
```

### Linux 二进制安装

以下命令以 x86_64 Linux 为例。ARM64 机器把 `linux_amd64` 换成 `linux_arm64`。

```bash
cd ~/Downloads
tar -xzf pxy_VERSION_linux_amd64.tar.gz
chmod +x pxy
sudo mv pxy /usr/local/bin/pxy
pxy init
```

重启终端，或执行 `pxy init` 输出的重新加载命令。之后即可使用：

```bash
pxy on
pxy status
pxy off
```

### Windows 二进制安装

1. 从 Releases 下载 `pxy_VERSION_windows_amd64.zip`。
2. 解压后得到 `pxy.exe`。
3. 在 `pxy.exe` 所在目录打开 PowerShell，执行：

```powershell
$InstallDir = "$env:LOCALAPPDATA\Programs\pxy"
New-Item -ItemType Directory -Force $InstallDir | Out-Null
Copy-Item .\pxy.exe "$InstallDir\pxy.exe" -Force

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($UserPath -split ";") -notcontains $InstallDir) {
  [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
  $env:Path = "$env:Path;$InstallDir"
}
```

确认安装成功：

```powershell
pxy init
```

`pxy init` 会写入 PowerShell profile。完成后重新打开 PowerShell，或执行输出中的 `. '...Microsoft.PowerShell_profile.ps1'` 命令。之后即可使用：

```powershell
pxy on
pxy status
pxy off
```

也可以从源码安装：

```bash
go install github.com/hczs/pxy@latest
```

Release 构建支持 macOS、Linux、Windows 的 amd64 / arm64。

## 快速开始

初始化：

```bash
pxy init
```

按提示选择自动检测到的代理配置，或手动输入代理地址。完成后重启终端，或执行命令输出中的重新加载命令。

开启代理：

```bash
pxy on
```

查看状态：

```bash
pxy status
```

测试当前出口 IP：

```bash
pxy test
```

查看版本：

```bash
pxy version
```

检查更新：

```bash
pxy update --check
```

更新到 GitHub Releases 最新版本：

```bash
pxy update
```

Windows 下运行中的 `pxy.exe` 不能直接覆盖；命令会下载新文件并提示手动替换路径。

关闭代理：

```bash
pxy off
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `pxy init` | 检测 shell 和代理配置，保存配置并安装 shell function |
| `pxy on` | 在当前 shell 中开启代理环境变量 |
| `pxy off` | 恢复或清理当前 shell 中的代理环境变量 |
| `pxy status` | 查看当前代理环境变量状态 |
| `pxy test` | 通过 `https://ipwho.is/` 测试当前代理 |
| `pxy list` | 列出检测到的本地代理软件 |
| `pxy config` | 手动重新配置代理 |
| `pxy version` | 查看构建版本 |
| `pxy update` | 检查或安装 GitHub Releases 最新版本 |

## 支持的环境变量

`pxy on` 会根据配置设置以下变量：

- `http_proxy` / `HTTP_PROXY`
- `https_proxy` / `HTTPS_PROXY`
- `all_proxy` / `ALL_PROXY`

HTTP 与 HTTPS 代理使用 `http://host:port`，SOCKS5 代理使用 `socks5://host:port`。

## 工作方式

普通 CLI 子进程无法直接修改父 shell 的环境变量，所以 `pxy init` 会向 shell profile 写入一个 `pxy` shell function。

之后执行 `pxy on` / `pxy off` 时，shell function 会调用 `pxy` 生成环境变量脚本，并在当前 shell 中执行。

## 配置

默认配置文件：

```text
~/.pxy/config.yaml
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
