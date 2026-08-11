# NOFX Calculation & Indicator Microservices Runtime (`nofx-calc-runtimes`)

本仓库承载 NOFX 系统中不常变动、包含重型依赖（C/Python 算法计算库、Node.js 交易所 SDK）的算力与技术指标微服务：

1. **`ta-lib`**: TA-Lib C 库与 Python 指标计算微服务
2. **`pandas-ta`**: Pandas-TA / Numpy / Scipy 技术分析算力微服务
3. **`trendln`**: Trendline 趋势线支撑/阻力位算法微服务
4. **`ccxt`**: CCXT 交易所数据与行情连接器微服务

## 镜像构建与发布

本仓库会在推送到 `main` 分支时通过 GitHub Actions 自动打包构建，发布至 GHCR (GitHub Container Registry)：

- `ghcr.io/pixel-flow-intelligence/nofx-talib:latest`
- `ghcr.io/pixel-flow-intelligence/nofx-pandas-ta:latest`
- `ghcr.io/pixel-flow-intelligence/nofx-trendln:latest`
- `ghcr.io/pixel-flow-intelligence/nofx-ccxt:latest`

主仓库 `nofx` 部署时直接引用上述 GHCR 镜像，无需在主仓库中重复构建这些重型算力镜像。
