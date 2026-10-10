# FXVIA 前端定制层

本目录保存 FXVIA 品牌资产、首页和登录布局的定制副本。升级与生产发布统一遵循私有运维仓库的 ops/UPGRADE.md；本目录不维护第二套发布流程。

## 资产与恢复工具

- components/：HomeView、AuthLayout、logo 的定制副本。
- locales/：首页中英文文案。
- seo/：入口 HTML、robots、sitemap 和站点验证文件。
- brand.css：品牌样式。
- apply_fxvia_overlay.py：将映射中的定制文件复制到 frontend。

Git 历史和备份有助恢复，但不保证合并后定制或新版修复自动正确保留。

## 升级时如何使用

1. 在升级分支合并固定上游 tag/commit，检查受影响定制。
2. 比较定制副本、当前目标文件与新版上游，尤其 AuthLayout、index.html 和页面逻辑。先适配上游行为及安全修复。
3. 只恢复确有需要的文件；脚本会覆盖全部已有映射源，不应合并后无条件运行。全部映射均经审查、确需恢复时才执行：

```bash
python front-fxvia/apply_fxvia_overlay.py
```

4. 审查最终 diff，运行受影响测试与构建。脚本输出成功只代表复制完成，不代表兼容性或安全性通过。
5. 上游升级自带的界面变化已有用户长期授权，无需重复看样；核对并保留自制主页与品牌。主动修改本地定制界面仍先看样，再一次性提交与构建。发布固定 digest，只部署已经验收的候选产物。

默认保留已发布历史，不 rebase/强推 main。CI 不可用时使用运维规范规定的替代验证；不以无分支保护、无 Actions 检查为由跳过测试或扫描。
