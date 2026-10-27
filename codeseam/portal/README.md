# CodeSeam 用户服务门户

这是 CodeSeam 自有的独立 Vue 3 + TypeScript + Vite 用户门户。它只面向普通用户，和原 Sub2API `frontend/` 运营后台分开构建、发布。

## 运行

```bash
pnpm install
pnpm dev
pnpm build
```

默认请求同源 `/api/v1`。部署到独立域名时可通过 `VITE_API_BASE_URL` 指向服务端 API。Windows 安装包地址通过 `VITE_WINDOWS_DOWNLOAD_URL` 配置；未配置时页面明确显示“暂未发布”，不会生成虚假链接。

P5 生产路径为 `/portal/`。可由现有 Go 主服务直接提供静态文件与安装包，无需 Vite 常驻。构建、部署路径和真实验收入口见 `../deploy/DELIVERY.md`。Logo 与客户端使用相同的正式 SVG。

## 复用的真实接口

- `POST /auth/login`、`POST /auth/register`、`POST /auth/refresh`、`POST /auth/logout`
- `POST /auth/forgot-password`
- `GET /auth/me`、`GET /user/profile`
- `GET /codeseam/service/current`
- `GET /usage/stats`（支持 `start_date`、`end_date`）
- `GET /usage/dashboard/trend`（按日趋势）

所有服务、订阅、Key 和用量数据以 Sub2API 当前用户 JWT 的返回为准。门户不调用客户端专用的 `/codeseam/client/activate` 或 `/codeseam/client/network-config`，也不接收或保存服务 Key、代理密码及上游凭据。
