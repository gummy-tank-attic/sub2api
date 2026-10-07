# FXVIA 前端品牌定制层与上游升级保护规范

本目录专门用于隔离与持久化保护 **FXVIA** 的全套前端定制设计与品牌资产，实现与官方开源上游 `sub2api` 的解耦。

---

## 🛡️ 为什么升级 sub2api 时主页绝对不会丢失？

1. **独立 Git 分支隔离**：
   - 本项目运行在 `gummy-tank-attic/sub2api` 独立仓库的 `main` 分支。
   - 所有定制化均通过原子 Commit 提交历史永久留存。
   - 上游更新时，通过标准 `git merge upstream/main` 进行功能同步，Git 会自动保留定制提交，绝不会强行覆写。

2. **独立生产镜像隔离**：
   - 生产环境采用专属容器镜像构建流水线：`ghcr.io/gummy-tank-attic/sub2api-fxvia@sha256:...`。
   - 独立镜像带有完整编译后的 FXVIA 定制页面，不受官方公共镜像变动影响。

3. **物理目录持久备份**：
   - `front-fxvia/components/HomeView.fxvia.vue`（定制极简居中主页）
   - `front-fxvia/components/AuthLayout.fxvia.vue`（登录/注册鉴权页官方红标）
   - `front-fxvia/components/logo.fxvia.svg`（FXVIΛ 官方红色纯字标矢量图）
   - `front-fxvia/locales/`（中英文双语高转化国际化文案字典）
   - `front-fxvia/brand.css`（全局品牌样式层）

4. **一键同步与恢复工具**：
   - 在任何升级、合并冲突或重置后，只需执行以下命令即可一秒还原全部品牌资产：
   ```bash
   python front-fxvia/apply_fxvia_overlay.py
   ```

---

## 🚀 标准上游平滑升级步骤

1. **拉取上游代码**：
   ```bash
   git fetch upstream
   git merge upstream/main
   ```
2. **应用并检查 FXVIA 定制层**：
   ```bash
   python front-fxvia/apply_fxvia_overlay.py
   ```
3. **本地/测试构建验证**：
   ```bash
   cd frontend && npm run build
   cd ../backend && go test ./...
   ```
4. **提交并推送至专属分支**：
   ```bash
   git add .
   git commit -m "chore(upstream): sync upstream and restore FXVIA brand overlay"
   git push origin main
   ```
5. **部署生效**：
   GitHub Actions 自动完成新镜像构建，部署时按明确镜像 digest 更新服务器，平滑无缝。
