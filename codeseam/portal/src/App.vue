<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import logo from './assets/codeseam-logo.svg'
import * as client from './api'
import type { Service, TrendPoint, UsageStats, User } from './api'

type Page = 'service' | 'usage' | 'help'
const page = ref<Page>('service')
const authMode = ref<'login' | 'register' | 'forgot'>('login')
const email = ref('')
const password = ref('')
const invitationCode = ref('')
const busy = ref(false)
const loading = ref(false)
const error = ref('')
const notice = ref('')
const user = ref<User | null>(null)
const services = ref<Service[]>([])
const stats = ref<UsageStats | null>(null)
const trend = ref<TrendPoint[]>([])
const startDate = ref(daysAgo(29))
const endDate = ref(today())
const downloadUrl = import.meta.env.VITE_WINDOWS_DOWNLOAD_URL || ''
const supportUrl = import.meta.env.VITE_SUPPORT_URL || ''

function today() { return new Date().toISOString().slice(0, 10) }
function daysAgo(days: number) { const date = new Date(); date.setDate(date.getDate() - days); return date.toISOString().slice(0, 10) }
function formatDate(value?: string | null) { return value ? new Date(value).toLocaleDateString('zh-CN') : '未提供' }
function formatNumber(value?: number | null) { return typeof value === 'number' ? new Intl.NumberFormat('zh-CN').format(value) : '未提供' }
function daysRemaining(value?: string | null) { if (!value) return null; return Math.max(0, Math.ceil((new Date(value).getTime() - Date.now()) / 86400000)) }
function stateLabel(state: string) { return ({ ACTIVE: '服务正常', DISABLED: '已停用', EXPIRED: '已到期', SUSPENDED: '已暂停', QUOTA_EXHAUSTED: '额度不足', NOT_STARTED: '尚未开始', NOT_PROVISIONED: '待运营开通', NOT_SUPPORTED: '暂不支持' } as Record<string, string>)[state] || state }
function stateClass(state: string) { return state === 'ACTIVE' ? 'active' : 'warning' }
const primaryService = computed(() => services.value[0])
const trendMax = computed(() => Math.max(1, ...trend.value.map((item) => item.total_tokens || 0)))
const totalQuota = computed(() => primaryService.value?.quota)
const usedQuota = computed(() => primaryService.value?.quota_used)
const quotaRatio = computed(() => totalQuota.value && totalQuota.value > 0 ? Math.min(100, ((usedQuota.value || 0) / totalQuota.value) * 100) : null)

async function refreshData() {
  if (!client.hasAuth()) { user.value = null; return }
  loading.value = true; error.value = ''
  try {
    user.value = await client.me()
    const current = await client.currentService()
    services.value = current.services || []
    if (page.value === 'usage') await loadUsage()
  } catch (err) { error.value = messageOf(err, '无法读取服务数据，请稍后重试') }
  finally { loading.value = false }
}
async function loadUsage() {
  if (startDate.value > endDate.value) { error.value = '开始日期不能晚于结束日期'; return }
  try {
    const [summary, daily] = await Promise.all([client.usageStats(startDate.value, endDate.value), client.usageTrend(startDate.value, endDate.value)])
    stats.value = summary; trend.value = daily.trend || []
  } catch (err) { error.value = messageOf(err, '无法读取用量数据，请稍后重试') }
}
async function submitAuth() {
  busy.value = true; error.value = ''; notice.value = ''
  try {
    if (authMode.value === 'login') await client.login(email.value, password.value)
    else if (authMode.value === 'register') await client.register(email.value, password.value, invitationCode.value)
    else { const result = await client.forgotPassword(email.value); notice.value = result.message || '如果邮箱存在，重置说明已发送'; busy.value = false; return }
    authMode.value = 'login'; await refreshData()
  } catch (err) { error.value = messageOf(err, '认证失败，请检查输入或服务端配置') }
  finally { busy.value = false }
}
async function signOut() { await client.logout(); user.value = null; services.value = []; stats.value = null; notice.value = '已安全退出' }
async function selectPage(next: Page) { page.value = next; error.value = ''; if (next === 'usage' && !stats.value) await loadUsage() }
function messageOf(err: unknown, fallback: string) { const candidate = err as { response?: { data?: { message?: string } }; message?: string }; return candidate?.response?.data?.message || candidate?.message || fallback }

onMounted(refreshData)
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <a class="brand" href="#" @click.prevent="selectPage('service')"><img class="brand-mark" :src="logo" alt="CodeSeam" /><span>CodeSeam</span></a>
      <div v-if="user" class="top-actions"><span class="user-email">{{ user.email || user.username }}</span><button class="ghost-button" @click="signOut">退出</button></div>
    </header>

    <main v-if="!user" class="auth-layout">
      <section class="intro"><p class="eyebrow">用户服务门户</p><h1>把服务状态和用量，<br><strong>清楚地交给你。</strong></h1><p>登录后查看本人服务期限、真实计量用量和客户端使用说明。所有信息都来自 CodeSeam 服务端实时接口。</p><div class="intro-points"><span>✓ 服务期限实时同步</span><span>✓ 用量按日期筛选</span><span>✓ 不展示服务 Key 和代理凭据</span></div></section>
      <section class="auth-card card">
        <div class="tabs"><button :class="{ selected: authMode === 'login' }" @click="authMode = 'login'">登录</button><button :class="{ selected: authMode === 'register' }" @click="authMode = 'register'">注册</button><button :class="{ selected: authMode === 'forgot' }" @click="authMode = 'forgot'">找回密码</button></div>
        <form @submit.prevent="submitAuth"><label>邮箱<input v-model.trim="email" type="email" autocomplete="email" required placeholder="name@example.com"></label><label v-if="authMode !== 'forgot'">密码<input v-model="password" type="password" :autocomplete="authMode === 'login' ? 'current-password' : 'new-password'" minlength="6" required placeholder="请输入密码"></label><label v-if="authMode === 'register'">邀请码（如需）<input v-model.trim="invitationCode" placeholder="运营提供时填写"></label><button class="primary-button wide" :disabled="busy">{{ busy ? '处理中…' : authMode === 'login' ? '登录门户' : authMode === 'register' ? '创建账号' : '发送找回说明' }}</button></form>
        <p v-if="notice" class="notice">{{ notice }}</p><p v-if="error" class="error">{{ error }}</p><p class="auth-note">认证由 Sub2API 原生 Auth API 处理，门户不会创建第二套账号体系。</p>
      </section>
    </main>

    <main v-else class="portal-layout">
      <aside class="sidebar"><p class="eyebrow">我的 CodeSeam</p><nav><button :class="{ current: page === 'service' }" @click="selectPage('service')"><span>▣</span> 我的服务</button><button :class="{ current: page === 'usage' }" @click="selectPage('usage')"><span>◒</span> 用量中心</button><button :class="{ current: page === 'help' }" @click="selectPage('help')"><span>?</span> 下载与帮助</button></nav><div class="sidebar-foot">服务状态和额度以服务端返回为准<br>刷新时间：{{ new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) }}</div></aside>
      <section class="content"><div class="content-head"><div><p class="eyebrow">{{ page === 'service' ? '服务概览' : page === 'usage' ? '真实计量' : '开始使用' }}</p><h1>{{ page === 'service' ? '我的服务' : page === 'usage' ? '用量中心' : '下载与帮助' }}</h1></div><button v-if="page !== 'help'" class="refresh-button" :disabled="loading" @click="refreshData">↻ {{ loading ? '刷新中' : '刷新' }}</button></div>
        <p v-if="error" class="error banner">{{ error }}</p>

        <template v-if="page === 'service'"><div v-if="!primaryService" class="empty card"><div class="empty-icon">○</div><h2>尚未开通 CodeSeam 服务</h2><p>当前账号没有可展示的服务。请联系运营开通或续费，开通后刷新本页面即可看到最新状态。</p><a v-if="supportUrl" class="primary-button" :href="supportUrl" target="_blank" rel="noreferrer">联系运营</a></div><div v-else class="service-grid"><article class="service-card card"><div class="service-card-head"><div><p class="muted">当前服务</p><h2>{{ primaryService.name || 'Claude Code Dedicated' }}</h2></div><span class="status-pill" :class="stateClass(primaryService.state)">{{ stateLabel(primaryService.state) }}</span></div><div class="service-dates"><div><span>开始日期</span><strong>{{ formatDate(primaryService.starts_at) }}</strong></div><div><span>到期日期</span><strong>{{ formatDate(primaryService.expires_at) }}</strong></div><div><span>剩余有效期</span><strong>{{ daysRemaining(primaryService.expires_at) === null ? '未提供' : `${daysRemaining(primaryService.expires_at)} 天` }}</strong></div></div><div v-if="quotaRatio !== null" class="quota"><div class="quota-label"><span>套餐额度</span><b>{{ formatNumber(usedQuota) }} / {{ formatNumber(totalQuota) }}</b></div><div class="progress"><i :style="{ width: `${quotaRatio}%` }"></i></div></div><p class="service-explain">{{ primaryService.state === 'ACTIVE' ? '服务已由运营配置，可使用 CodeSeam 客户端连接。' : '当前服务不可用，请联系运营确认授权、有效期或额度状态。' }}</p></article><article class="side-stack"><div class="mini-card card"><span class="mini-icon">↓</span><div><h3>客户端</h3><p>下载 Windows 客户端并登录你的 CodeSeam 账号。</p></div><a v-if="downloadUrl" class="text-link" :href="downloadUrl" target="_blank" rel="noreferrer">下载客户端 →</a><span v-else class="unavailable">安装包暂未发布</span></div><div class="mini-card card"><span class="mini-icon">↗</span><div><h3>需要续费？</h3><p>人工续费和服务异常请联系运营，更新后刷新即可生效。</p></div><a v-if="supportUrl" class="text-link" :href="supportUrl" target="_blank" rel="noreferrer">联系运营 →</a><span v-else class="unavailable">请联系你的服务提供方</span></div></article></div></template>

        <template v-else-if="page === 'usage'"><div class="filters card"><label>开始日期<input v-model="startDate" type="date"></label><span>至</span><label>结束日期<input v-model="endDate" type="date"></label><button class="primary-button" @click="loadUsage">查询用量</button></div><div class="metric-grid"><div class="metric card"><span>总请求</span><strong>{{ formatNumber(stats?.total_requests) }}</strong></div><div class="metric card"><span>总 Token</span><strong>{{ formatNumber(stats?.total_tokens) }}</strong></div><div class="metric card"><span>Input Token</span><strong>{{ formatNumber(stats?.total_input_tokens) }}</strong></div><div class="metric card"><span>Output Token</span><strong>{{ formatNumber(stats?.total_output_tokens) }}</strong></div><div class="metric card"><span>Cache Token</span><strong>{{ formatNumber(stats?.total_cache_tokens ?? stats?.total_cache_creation_tokens ?? stats?.total_cache_read_tokens) }}</strong></div><div class="metric card"><span>实际扣除额度</span><strong>{{ typeof stats?.total_actual_cost === 'number' ? stats.total_actual_cost.toFixed(4) : '未提供' }}</strong></div></div><div class="chart-card card"><div class="card-heading"><div><h2>每日趋势</h2><p>按服务端计量数据展示，空数据不会被推算。</p></div><span class="muted">{{ startDate }} — {{ endDate }}</span></div><div v-if="trend.length" class="bars"><div v-for="item in trend" :key="item.date" class="bar-item"><div class="bar-value">{{ formatNumber(item.total_tokens) }}</div><div class="bar" :style="{ height: `${Math.max(4, ((item.total_tokens || 0) / trendMax) * 150)}px` }"></div><small>{{ item.date.slice(5) }}</small></div></div><div v-else class="empty-chart">当前日期范围暂无趋势数据</div><p class="semantic-note">Sub2API Token 计量与 Claude 官方 5h/1w 用量窗口不是同一概念，本页不提供官方窗口剩余百分比。</p></div></template>

        <template v-else><div class="help-grid"><article class="help-card card"><span class="help-number">01</span><h2>下载并安装</h2><p>从运营提供的正式地址下载 Windows 客户端，安装后使用本门户的账号登录。安装包发布前，下载入口会明确显示暂未发布。</p><a v-if="downloadUrl" class="primary-button" :href="downloadUrl" target="_blank" rel="noreferrer">下载 Windows 客户端</a></article><article class="help-card card"><span class="help-number">02</span><h2>连接 Claude Code</h2><p>客户端会检查本地配置并管理连接路径。按客户端提示完成检查后，再启动 Claude Code。不要在门户或聊天中粘贴服务 Key、代理密码或上游账号凭据。</p></article><article class="help-card card"><span class="help-number">03</span><h2>遇到连接异常</h2><p>先在客户端执行连接诊断，确认客户端能联系服务端且服务仍在有效期内。网络保护和非 API 流量的覆盖范围以客户端真实诊断结果为准；门户不宣称全面防泄漏。</p><a v-if="supportUrl" class="text-link" :href="supportUrl" target="_blank" rel="noreferrer">联系运营 →</a></article></div></template>
      </section>
    </main>
  </div>
</template>
