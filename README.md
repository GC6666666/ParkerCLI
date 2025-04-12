# ParkerCli

一款面向 Go 后端开发调试、部署、发布的全能 CLI 工具，基于 urfave/cli 构建。

## 功能特性

- **debug**: 调试信息、日志过滤、服务信息查看等
- **run**: 启动后端服务、跑特定任务或Job
- **config**: 初始化、查看、更新项目配置
- **build**: 一键触发编译，包括代码、容器镜像
- **test**: 整合单元测试与集成测试
- **logs**: 日志聚合与过滤
- **migrate**: 数据库迁移、版本回退
- **release**: 版本打包、发布
- **version**: 查看当前版本

## 安装

```bash
# 克隆仓库
git clone https://github.com/parker/ParkerCli.git

# 进入项目目录
cd ParkerCli

# 安装依赖
go mod tidy

# 构建
go build -o ParkerCli
```

## 使用示例

1. 查看帮助

```bash
./ParkerCli --help
./ParkerCli debug --help
```

2. 调试日志

```bash
./ParkerCli debug logs --level=ERROR
```

3. 启动服务

```bash
./ParkerCli run server
```

4. 生成 Docker 镜像

```bash
./ParkerCli build image
```

5. 查看版本

```bash
./ParkerCli version
```

## 命令详解

### debug 命令

```bash
# 查看所有日志
./ParkerCli debug logs

# 按日志级别过滤
./ParkerCli debug logs --level=ERROR

# 查看服务信息
./ParkerCli debug info

# 测试API接口
./ParkerCli debug test --api=/health
```

### run 命令

```bash
# 启动服务
./ParkerCli run server

# 执行一次性任务
./ParkerCli run job --once
```

### config 命令

```bash
# 初始化配置
./ParkerCli config init

# 查看配置
./ParkerCli config show

# 设置配置项
./ParkerCli config set --key=DB_HOST --value=127.0.0.1
```

### build 命令

```bash
# 编译代码
./ParkerCli build code

# 构建Docker镜像
./ParkerCli build image
```

### test 命令

```bash
# 运行单元测试
./ParkerCli test unit

# 运行集成测试
./ParkerCli test integration --setup
```

### logs 命令

```bash
# 实时跟踪日志
./ParkerCli logs tail

# 过滤日志内容
./ParkerCli logs grep --keyword="error"
```

### migrate 命令

```bash
# 创建新迁移
./ParkerCli migrate create --name="create_user_table"

# 执行数据库升级
./ParkerCli migrate up

# 数据库版本回退
./ParkerCli migrate down

# 查看迁移状态
./ParkerCli migrate status
```

### release 命令

```bash
# 构建发行版
./ParkerCli release build --version=0.1.0

# 发布版本
./ParkerCli release publish --tag=v0.1.0
```

## 项目结构

```
ParkerCli/
├─ cmd/               # 命令定义与处理
│  ├─ debug.go        # 调试命令
│  ├─ run.go          # 运行命令
│  ├─ config.go       # 配置命令
│  ├─ build.go        # 构建命令
│  ├─ test.go         # 测试命令
│  ├─ logs.go         # 日志命令
│  ├─ migrate.go      # 迁移命令
│  └─ release.go      # 发布命令
├─ internal/          # 内部业务逻辑
│  ├─ config/         # 配置管理
│  ├─ debug/          # 调试工具
│  ├─ runner/         # 服务运行器
│  ├─ migrator/       # 数据库迁移
│  ├─ builder/        # 构建工具
│  └─ utils/          # 通用工具函数
├─ pkg/               # 可重用公共库
│  ├─ logger/         # 日志库
│  └─ httpclient/     # HTTP客户端
├─ test/              # 测试相关
│  ├─ fixtures/       # 测试数据
│  └─ mocks/          # 模拟对象
├─ docs/              # 文档
│  ├─ developLog.md          # 开发日志
│  ├─ developProcess.md      # 开发流程
│  ├─ project_architecture.md # 项目架构文档
│  └─ docV1.0.md             # 功能文档
├─ scripts/           # 辅助脚本
│  ├─ autocomplete.sh        # Bash自动补全
│  └─ autocomplete.ps1       # PowerShell自动补全
├─ migrations/        # 迁移文件存储
├─ main.go            # 程序入口
└─ README.md          # 项目说明
```

## 架构说明

ParkerCli采用模块化、分层架构设计，主要分为命令层(cmd)、业务逻辑层(internal)和基础设施层(pkg)。

### 设计原则

- **关注点分离**：每个模块职责单一，降低耦合
- **接口优先**：通过接口定义模块边界，便于测试和替换实现
- **依赖注入**：显式管理组件依赖关系
- **易扩展性**：预留扩展点，支持添加新命令和功能

### 模块使用情况

目前项目中已实现的功能：
- **cmd/debug.go** 使用了 **internal/debug** 模块，实现了调试功能
- **cmd/config.go** 使用了 **internal/config** 模块，实现了配置管理功能
- **cmd/build.go** 使用了 **internal/builder** 模块，实现了构建功能
- **cmd/migrate.go** 使用了 **internal/migrator** 模块，实现了数据库迁移功能
- 其他命令模块正在逐步重构，迁移到对应的internal实现中

项目提供了详细的架构流程图和函数关系图，帮助开发者快速理解系统：

- **命令执行流程图**：展示从用户输入到结果输出的完整路径
- **模块依赖关系图**：清晰呈现不同包和模块间的调用关系
- **时序图**：详细展示主要命令（debug, config, build, migrate, logs, run, test, release）的执行过程
- **类图**：展示接口与实现的关系

所有架构图表都使用Mermaid语法创建，集成在[项目架构文档](docs/project_architecture.md)中。

## 拓展开发

ParkerCli设计为可扩展的命令行工具，你可以通过以下方式进行拓展：

1. 在cmd目录下创建新的命令文件
2. 在internal目录下添加具体的业务逻辑
3. 在main.go中注册新命令

详细的架构设计和扩展方法请参考 [项目架构文档](docs/project_architecture.md)。

## 测试结果

最新的功能测试显示ParkerCli运行良好，主要测试结果如下：

1. **基础命令**：所有命令和子命令结构清晰，帮助信息完整
2. **debug模块**：日志过滤、服务信息查看和API测试功能正常
3. **config模块**：配置初始化、查看和修改功能工作正常
4. **run模块**：服务启动和任务执行功能正常
5. **build模块**：代码构建功能工作正常，单元测试已完成验证构建Go程序和Docker镜像功能
6. **test模块**：单元测试执行功能正常
7. **migrate模块**：数据库迁移管理功能完善
8. **release模块**：多平台构建支持，但在Windows环境可能有交互问题
9. **version命令**：版本信息展示正确

详细的测试记录请查看 `docs/developProcess.md` 和 `docs/developLog.md`。

## 命令自动补全

ParkerCli提供了命令自动补全功能，支持Bash和PowerShell环境：

### Bash环境

```bash
# 临时启用自动补全
source scripts/autocomplete.sh

# 永久启用自动补全
echo 'source /path/to/ParkerCli/scripts/autocomplete.sh' >> ~/.bashrc
```

### PowerShell环境

```powershell
# 临时启用自动补全
. .\scripts\autocomplete.ps1

# 永久启用自动补全
# 编辑PowerShell配置文件
notepad $PROFILE
# 添加以下行
. C:\path\to\ParkerCli\scripts\autocomplete.ps1
```

## 最佳实践

1. **配置管理**
   - 先运行 `config init` 初始化配置
   - 使用 `config set` 只修改需要的配置项
   - 配置更改后重启相关服务

2. **服务管理**
   - 开发环境使用 `run server` 启动服务
   - 生产环境建议使用 `run server --release` 启动优化后的服务

3. **调试技巧**
   - 使用 `debug info` 快速查看系统状态
   - 问题定位时结合 `debug logs` 和 `logs grep` 筛选日志

4. **数据库管理**
   - 始终使用 `migrate status` 检查当前数据库状态
   - 数据库更改前做好备份

## 已知问题

1. **Windows环境**：
   - release build命令在Windows PowerShell中可能有光标定位问题
   - 某些依赖外部工具的功能需要确保工具已安装

2. **Docker相关**：
   - build image和其他Docker相关功能需要本地安装Docker
   - 容器功能需要在具有Docker环境的系统上测试

3. **Git集成**：
   - release publish功能依赖于Git，需要确保本地存在有效的Git仓库

## 开发路线图

1. **近期计划**：
   - 修复Windows环境下的交互问题
   - 增强错误处理和友好提示
   - 添加自动补全功能

2. **中期计划**：
   - 增加插件系统支持扩展命令
   - 完善自动化测试覆盖率
   - 添加交互式命令模式

3. **长期愿景**：
   - 构建完整的Web界面管理功能
   - 支持更多的云平台集成
   - 建立插件生态系统

详细的开发计划和进展请参考 [开发日志](docs/developLog.md)。

## 依赖库

- [github.com/urfave/cli/v2](https://github.com/urfave/cli) - CLI框架
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - Web框架
- [github.com/robfig/cron/v3](https://github.com/robfig/cron) - 定时任务
- [github.com/spf13/viper](https://github.com/spf13/viper) - 配置管理
- [github.com/docker/docker](https://github.com/docker/docker) - Docker API

## 贡献指南

欢迎提交问题报告和拉取请求来帮助改进ParkerCli。在提交贡献前，请阅读我们的[架构文档](docs/project_architecture.md)了解设计理念。

## 许可证

MIT 