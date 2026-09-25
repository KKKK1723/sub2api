# KeepCoding API 首页预览

独立的 Vue 首页预览。首屏突出模型 API、AI 订阅和接码服务，品牌保留在导航与页脚。包含中英文、深浅主题、动态流线背景、套餐筛选、接口示例复制及微信咨询弹窗。

## 运行与构建

在 `frontend` 目录、已安装项目依赖的环境中运行：

```powershell
node node_modules/vite/bin/vite.js taokc-home-preview
```

默认地址为 `http://127.0.0.1:3180/`，端口被占用时 Vite 自动选择下一个可用端口。

本地审核入口：

- `/`：首页。
- `/login`：登录页，包含密码显隐、字段校验和账号找回联系方式。
- `/register`：注册页，包含填写信息、邮箱验证两个步骤。

登录和注册复用首页的模型图标、流线背景、配色与主题语言偏好。账号表单仅展示本地交互，不调用线上账号接口、不发送邮件、不保存密码或验证码。

```powershell
node node_modules/vite/bin/vite.js build taokc-home-preview
```

构建结果位于 `taokc-home-preview/dist`。此目录独立于正式前端入口，当前只用于本地预览，未部署生产站。

## 内容与入口

- `src/content.js` 集中维护双语文案、套餐价格及联系方式。套餐价格与质保说明来自现有 `src/views/user/AiProductServiceView.vue`。
- 首页控制台按钮进入本地 `/login`；密钥与可用渠道入口指向 `https://taokc.xyz`；文档沿用公开站点配置 `https://1003gou.sbs`。
- 订阅和接码通过微信人工确认。接码区域展示手机号验证流程图与联系入口，没有下单表单、支付或自动接码接口。
- 二维码复用 `frontend/public/contact/wechat.png`；基础图标复用项目的 `Icon.vue`。
- `public/brands` 的模型品牌图标来自 LobeHub Icons，MIT 许可见同目录 `LICENSE.txt`；品牌名称和标识属于对应权利人。
- `src/style.css` 与 `src/HeroFlow.vue` 为本预览的样式与 Canvas 背景实现。
- 首页文字与内容在进入视口时依次显现；`src/ToolsMarquee.vue` 提供可暂停的工具循环流动带，离开视口或切到后台时暂停，并尊重系统的减少动态效果设置。`src/SmsFlow.vue` 展示接码流程。
- `src/AuthPreview.vue`、`src/auth.css` 和 `src/auth-content.js` 为账号页预览。按本次读取到的公开配置展示邮箱密码登录、QQ/Gmail 邮箱注册与邮箱验证，未启用的第三方登录、邀请码和优惠码字段不展示；正式接入时继续以服务端配置为准。

主题和语言分别保存在浏览器的 `keepcoding-theme`、`keepcoding-locale` 本地存储中。
