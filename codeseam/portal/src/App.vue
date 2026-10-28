<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import logo from './assets/codeseam-logo.svg'
import * as client from './api'
import type { Service, TrendPoint, UsageStats, User } from './api'

type Page = 'overview' | 'service' | 'usage' | 'devices' | 'history' | 'help'
type AuthMode = 'login' | 'register' | 'forgot'
const nav: { id: Page; label: string; icon: string; href: string; group: 'account' | 'support'; planned?: boolean }[] = [
  { id: 'overview', label: '首页', icon: 'dashboard', href: '/app', group: 'account' },
  { id: 'devices', label: '设备', icon: 'monitor', href: '/app/devices', group: 'account', planned: true },
  { id: 'history', label: '历史记录', icon: 'clock', href: '/app/history', group: 'account', planned: true },
  { id: 'service', label: '我的服务', icon: 'shield', href: '/app/service', group: 'account' },
  { id: 'usage', label: '用量分析', icon: 'chart', href: '/app/usage', group: 'account' },
  { id: 'help', label: '下载与帮助', icon: 'download', href: '/app/download', group: 'support' }
]
const pathPage: Record<string, Page> = {
  '/app': 'overview', '/app/service': 'service', '/app/usage': 'usage',
  '/app/devices': 'devices', '/app/history': 'history', '/app/download': 'help'
}
function routePage(): Page { return pathPage[window.location.pathname.replace(/\/$/, '')] || 'overview' }
function navigate(href: string) {
  window.history.pushState(null, '', href)
  window.dispatchEvent(new PopStateEvent('popstate'))
}
const page = ref<Page>(routePage())
const authReady = ref(false)
const menuOpen = ref(false)
const sidebarCollapsed = ref(false)
const accountMenuOpen = ref(false)
const overviewRangeDays = ref(30)
const activityTrend = ref<TrendPoint[]>([])
const activityLoading = ref(false)
const activityLoaded = ref(false)
const activityError = ref('')
const authMode = ref<AuthMode>('login')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const passwordVisible = ref(false)
const confirmPasswordVisible = ref(false)
const invitationCode = ref('')
const busy = ref(false)
const loading = ref(false)
const usageLoading = ref(false)
const error = ref('')
const notice = ref('')
const user = ref<User | null>(null)
const services = ref<Service[]>([])
const stats = ref<UsageStats | null>(null)
const trend = ref<TrendPoint[]>([])
const startDate = ref(daysAgo(29))
const endDate = ref(today())
const lastRefresh = ref('')
const theme = ref<'light' | 'dark'>(
  localStorage.getItem('cs.portal.theme') === 'dark' ? 'dark' : 'light'
)
const downloadUrl = import.meta.env.VITE_WINDOWS_DOWNLOAD_URL || ''
const supportUrl = import.meta.env.VITE_SUPPORT_URL || ''

watch(theme, (value) => {
  document.documentElement.dataset.theme = value
  localStorage.setItem('cs.portal.theme', value)
}, { immediate: true })

function localDate(value: Date) {
  return [value.getFullYear(), String(value.getMonth() + 1).padStart(2, '0'), String(value.getDate()).padStart(2, '0')].join('-')
}
function today() { return localDate(new Date()) }
function daysAgo(days: number) { const date = new Date(); date.setDate(date.getDate() - days); return localDate(date) }
function formatDate(value?: string | null) { return value ? new Date(value).toLocaleDateString('zh-CN') : '未提供' }
function formatNumber(value?: number | null) { return typeof value === 'number' ? new Intl.NumberFormat('zh-CN').format(value) : '未提供' }
function daysRemaining(value?: string | null) {
  if (!value) return null
  const remaining = new Date(value).getTime() - Date.now()
  return Number.isFinite(remaining) ? Math.max(0, Math.ceil(remaining / 86400000)) : null
}
function stateLabel(state: string) {
  return ({ ACTIVE: '服务正常', DISABLED: '已停用', EXPIRED: '已到期', SUSPENDED: '已暂停', QUOTA_EXHAUSTED: '额度不足', NOT_STARTED: '尚未开始', NOT_PROVISIONED: '待开通', NOT_SUPPORTED: '暂不支持' } as Record<string, string>)[state] || state
}
function messageOf(err: unknown, fallback: string) {
  const candidate = err as { response?: { data?: { message?: string } }; message?: string }
  return candidate?.response?.data?.message || candidate?.message || fallback
}
const primaryService = computed(() => services.value[0])
const serviceActive = computed(() => primaryService.value?.state === 'ACTIVE')
const userName = computed(() => user.value?.display_name || user.value?.email || user.value?.username || 'CodeSeam 用户')
const initials = computed(() => userName.value.slice(0, 1).toUpperCase())
const totalQuota = computed(() => primaryService.value?.quota)
const usedQuota = computed(() => primaryService.value?.quota_used)
const quotaRatio = computed(() => totalQuota.value && totalQuota.value > 0 ? Math.min(100, Math.max(0, ((usedQuota.value || 0) / totalQuota.value) * 100)) : null)
const remainingQuota = computed(() => totalQuota.value == null || usedQuota.value == null ? null : Math.max(0, totalQuota.value - usedQuota.value))
const dayCount = computed(() => daysRemaining(primaryService.value?.expires_at))
const trendMax = computed(() => Math.max(1, ...trend.value.map((item) => item.total_tokens || 0)))
const totalCache = computed(() => stats.value?.total_cache_tokens ?? (
  stats.value?.total_cache_creation_tokens != null || stats.value?.total_cache_read_tokens != null
    ? (stats.value?.total_cache_creation_tokens || 0) + (stats.value?.total_cache_read_tokens || 0)
    : undefined
))
const activityCells = computed(() => {
  const byDate = new Map(activityTrend.value.map(item => [item.date.slice(0, 10), item.total_tokens || 0]))
  const peak = Math.max(1, ...Array.from(byDate.values()))
  return Array.from({ length: 365 }, (_, index) => {
    const date = daysAgo(364 - index)
    const amount = byDate.get(date) || 0
    const ratio = amount / peak
    return { date, amount, level: !amount ? 0 : ratio < .2 ? 1 : ratio < .45 ? 2 : ratio < .7 ? 3 : 4 }
  })
})
const activeDays = computed(() => activityCells.value.filter(cell => cell.amount > 0).length)
function stat(value?: number | null) { return typeof value === 'number' ? formatNumber(value) : '—' }
function barHeight(item: TrendPoint) { return Math.max(3, Math.round(((item.total_tokens || 0) / trendMax.value) * 148)) }
function toggleTheme() { theme.value = theme.value === 'light' ? 'dark' : 'light' }

function syncAuthMode() {
  const mode = window.location.pathname.slice(1).toLowerCase()
  authMode.value = mode === 'register' || mode === 'forgot' ? mode : 'login'
  page.value = routePage()
  error.value = ''
  notice.value = ''
  passwordVisible.value = false
  confirmPasswordVisible.value = false
}
function switchAuthMode(mode: AuthMode) {
  if (busy.value) return
  navigate('/' + mode)
}
function handleRoute() {
  syncAuthMode()
  menuOpen.value = false
  accountMenuOpen.value = false
  if (user.value && ['/login', '/register', '/forgot'].includes(window.location.pathname)) navigate('/app')
  else if (!user.value && authReady.value && window.location.pathname.startsWith('/app')) {
    navigate('/login?next=' + encodeURIComponent(window.location.pathname))
  } else if (user.value && (page.value === 'overview' || page.value === 'usage') && !stats.value) void loadUsage()
  if (user.value && page.value === 'overview' && !activityLoaded.value) void loadActivity()
}


async function refreshData() {
  if (!client.hasAuth()) {
    user.value = null
    authReady.value = true
    handleRoute()
    return
  }
  loading.value = true
  error.value = ''
  try {
    user.value = await client.me()
    if (['/login', '/register', '/forgot'].includes(window.location.pathname)) navigate('/app')
    const current = await client.currentService()
    services.value = current.services || []
    lastRefresh.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    if (page.value === 'overview' || page.value === 'usage') await loadUsage()
    if (page.value === 'overview' && !activityLoaded.value) void loadActivity()
  } catch (err) {
    const status = (err as { response?: { status?: number } })?.response?.status
    if (!user.value && status === 401) client.clearAuth()
    error.value = messageOf(err, '暂时无法读取账户信息，请检查连接后重试')
  } finally {
    loading.value = false
    authReady.value = true
    if (!user.value && !client.hasAuth()) handleRoute()
  }
}
async function loadUsage() {
  if (startDate.value > endDate.value) { error.value = '开始日期不能晚于结束日期'; return }
  usageLoading.value = true
  error.value = ''
  try {
    const [summary, daily] = await Promise.all([
      client.usageStats(startDate.value, endDate.value),
      client.usageTrend(startDate.value, endDate.value)
    ])
    stats.value = summary
    trend.value = daily.trend || []
  } catch (err) { error.value = messageOf(err, '无法读取用量数据，请稍后重试') }
  finally { usageLoading.value = false }
}
async function setRange(days: number) {
  overviewRangeDays.value = days
  startDate.value = daysAgo(days - 1)
  endDate.value = today()
  await loadUsage()
}
async function refreshOverview() {
  await Promise.all([refreshData(), loadActivity(true)])
}
async function loadActivity(force = false) {
  if (activityLoading.value || (activityLoaded.value && !force)) return
  activityLoading.value = true
  activityError.value = ''
  try {
    const daily = await client.usageTrend(daysAgo(364), today())
    activityTrend.value = daily.trend || []
    activityLoaded.value = true
  } catch (err) {
    activityError.value = messageOf(err, '活动数据暂时无法读取')
  } finally {
    activityLoading.value = false
  }
}
async function submitAuth() {
  busy.value = true
  error.value = ''
  notice.value = ''
  if (authMode.value === 'register' && password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    busy.value = false
    return
  }
  try {
    if (authMode.value === 'login' || authMode.value === 'register') {
      const response = authMode.value === 'login'
        ? await client.login(email.value, password.value)
        : await client.register(email.value, password.value, invitationCode.value)
      user.value = response.user || await client.me()
    }
    else {
      const result = await client.forgotPassword(email.value)
      notice.value = result.message || '如果邮箱存在，重置说明已发送'
      return
    }
    authMode.value = 'login'
    navigate('/app')
    await refreshData()
  } catch (err) { error.value = messageOf(err, '认证失败，请检查输入或服务端配置') }
  finally { busy.value = false }
}
async function signOut() {
  await client.logout()
  user.value = null
  services.value = []
  stats.value = null
  trend.value = []
  page.value = 'overview'
  notice.value = ''
  error.value = ''
  authReady.value = true
  window.location.assign('/login')
}
async function selectPage(next: Page) {
  const target = nav.find((item) => item.id === next)
  if (!target) return
  navigate(target.href)
  error.value = ''
  if ((next === 'overview' || next === 'usage') && !stats.value) await loadUsage()
  if (next === 'overview' && !activityLoaded.value) void loadActivity()
}
onMounted(() => {
  handleRoute()
  window.addEventListener('popstate', handleRoute)
  void refreshData()
})
onUnmounted(() => window.removeEventListener('popstate', handleRoute))
</script>

<template>
  <svg class="svg-sprite" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <symbol id="i-menu" viewBox="0 0 24 24"><path d="M4 7h16M4 12h16M4 17h16"/></symbol>
    <symbol id="i-chevron-left" viewBox="0 0 24 24"><path d="m14 6-6 6 6 6"/></symbol>
    <symbol id="i-dashboard" viewBox="0 0 24 24"><path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z"/><path d="M9 21v-7h6v7"/></symbol>
    <symbol id="i-chart" viewBox="0 0 24 24"><path d="M4 20V10m6 10V4m6 16v-7m5 7H2"/></symbol>
    <symbol id="i-monitor" viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="14" rx="2"/><path d="M8 22h8m-4-4v4"/></symbol>
    <symbol id="i-download" viewBox="0 0 24 24"><path d="M12 3v12m-5-5 5 5 5-5M4 17v3h16v-3"/></symbol>
    <symbol id="i-moon" viewBox="0 0 24 24"><path d="M20 15.5A8.5 8.5 0 0 1 8.5 4 8.5 8.5 0 1 0 20 15.5Z"/></symbol>
    <symbol id="i-sun" viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M4.93 4.93l1.42 1.42m11.3 11.3 1.42 1.42M2 12h2m16 0h2M4.93 19.07l1.42-1.42m11.3-11.3 1.42-1.42"/></symbol>
    <symbol id="i-refresh" viewBox="0 0 24 24"><path d="M20 11a8 8 0 0 0-14.4-4.8L3 9m0-5v5h5m-4 4a8 8 0 0 0 14.4 4.8L21 15m0 5v-5h-5"/></symbol>
    <symbol id="i-arrow" viewBox="0 0 24 24"><path d="M5 12h14m-6-6 6 6-6 6"/></symbol>
    <symbol id="i-calendar" viewBox="0 0 24 24"><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M8 2v6m8-6v6M3 11h18"/></symbol>
    <symbol id="i-shield" viewBox="0 0 24 24"><path d="m12 2 9 4v6c0 5-3.5 8-9 10-5.5-2-9-5-9-10V6z"/><path d="m8 12 3 3 5-6"/></symbol>
    <symbol id="i-terminal" viewBox="0 0 24 24"><rect x="2" y="3" width="20" height="18" rx="2"/><path d="m6 8 4 4-4 4m6 0h6"/></symbol>
    <symbol id="i-help" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M9.6 9a2.5 2.5 0 1 1 4 2c-1 .8-1.6 1.1-1.6 3m0 3h.01"/></symbol>
    <symbol id="i-check" viewBox="0 0 24 24"><path d="m4 12 5 5L20 6"/></symbol>
    <symbol id="i-clock" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l4 2"/></symbol>
    <symbol id="i-mail" viewBox="0 0 24 24"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m3 6 9 7 9-7"/></symbol>
    <symbol id="i-log-out" viewBox="0 0 24 24"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4m7 4 5 5-5 5m5-5H9"/></symbol>
  </svg>

  <div class="portal-root" :class="{ 'cs-auth-root': authReady && !user, 'portal-dashboard': !!user, 'sidebar-is-collapsed': sidebarCollapsed }">
    <header v-if="!user" class="cs-auth-topbar">
      <a href="/" class="cs-auth-home" aria-label="CodeSeam 首页">
        <img :src="logo" alt="CodeSeam" />
      </a>
      <a href="/#workflow" class="cs-auth-docs">使用文档</a>
    </header>

    <header v-else-if="user" class="topbar">
      <div class="dashboard-status-line">
        <span class="dashboard-status-dot" :class="{ 'is-active': serviceActive }"></span>
        <strong>{{ primaryService ? stateLabel(primaryService.state) : '待开通' }}</strong>
        <span class="dashboard-status-divider">·</span>
        <span>CodeSeam 用户服务</span>
        <span v-if="lastRefresh" class="dashboard-status-update">更新于 {{ lastRefresh }}</span>
      </div>
      <div class="dashboard-toolbar">
        <div class="dashboard-toolbar-context">
          <button type="button" class="dashboard-mobile-menu" :aria-expanded="menuOpen" aria-label="打开菜单" @click="menuOpen = !menuOpen">
            <svg class="icon"><use href="#i-menu"/></svg>
          </button>
          <span class="toolbar-context-logo"><img :src="logo" alt=""/></span>
          <span class="toolbar-context-name" :title="userName">{{ userName }}</span>
          <span class="toolbar-context-type">个人账户</span>
        </div>
        <div class="dashboard-toolbar-actions">
          <a href="/" class="dashboard-toolbar-link">返回官网</a>
          <button type="button" class="icon-button dashboard-theme-button" :aria-label="theme === 'light' ? '切换暗色主题' : '切换浅色主题'" @click="toggleTheme">
            <svg class="icon"><use :href="theme === 'light' ? '#i-moon' : '#i-sun'" /></svg>
          </button>
          <div class="dashboard-account-area">
            <button type="button" class="dashboard-account-trigger" :aria-expanded="accountMenuOpen" aria-label="账户菜单" @click="accountMenuOpen = !accountMenuOpen">{{ initials }}</button>
            <div v-if="accountMenuOpen" class="dashboard-account-popover">
              <strong :title="userName">{{ userName }}</strong>
              <span>CodeSeam 个人账户</span>
              <a href="/app/service" @click.prevent="selectPage('service')">我的服务</a>
              <button type="button" @click="signOut">退出登录</button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <div v-if="!authReady" class="portal-auth-loading"><img :src="logo" alt="" /><p>正在确认登录状态…</p></div>
    <main v-if="authReady && !user" class="cs-auth-shell">
      <section class="cs-auth-card" :class="'cs-auth-' + authMode" aria-labelledby="cs-auth-title">
        <img :src="logo" alt="" class="cs-auth-mark" />
        <h1 id="cs-auth-title">{{ authMode === 'login' ? '登录你的账户' : authMode === 'register' ? '创建你的账户' : '找回密码' }}</h1>
        <p class="cs-auth-subtitle">{{ authMode === 'login' ? '使用你的 CodeSeam 账户继续。' : authMode === 'register' ? '创建 CodeSeam 账户，开始使用专属服务。' : '输入邮箱，我们将向你发送重置说明。' }}</p>

        <form class="cs-auth-form" @submit.prevent="submitAuth">
          <label class="cs-auth-field">
            <span class="cs-auth-field-label">邮箱地址</span>
            <input v-model.trim="email" name="email" type="email" autocomplete="email" placeholder="name@example.com" required />
          </label>

          <label v-if="authMode !== 'forgot'" class="cs-auth-field cs-auth-password-field">
            <span class="cs-auth-field-label">密码</span>
            <input v-model="password" name="password" :type="passwordVisible ? 'text' : 'password'" :autocomplete="authMode === 'login' ? 'current-password' : 'new-password'" placeholder="请输入密码" minlength="6" required />
            <button type="button" class="cs-auth-eye" :aria-label="passwordVisible ? '隐藏密码' : '显示密码'" :aria-pressed="passwordVisible" @click.prevent="passwordVisible = !passwordVisible">
              <svg v-if="!passwordVisible" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" /><circle cx="12" cy="12" r="3" /></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 3l18 18M10.6 5.6c.45-.07.92-.1 1.4-.1 6 0 9.5 6.5 9.5 6.5a17.9 17.9 0 0 1-3.7 4.1M6.2 6.9A17.5 17.5 0 0 0 2.5 12S6 18.5 12 18.5c1.04 0 2.03-.2 2.95-.54M10 10a3 3 0 0 0 4 4" /></svg>
            </button>
          </label>

          <label v-if="authMode === 'register'" class="cs-auth-field cs-auth-password-field">
            <span class="cs-auth-field-label">确认密码</span>
            <input v-model="confirmPassword" name="confirm-password" :type="confirmPasswordVisible ? 'text' : 'password'" autocomplete="new-password" placeholder="再次输入密码" minlength="6" required />
            <button type="button" class="cs-auth-eye" :aria-label="confirmPasswordVisible ? '隐藏确认密码' : '显示确认密码'" :aria-pressed="confirmPasswordVisible" @click.prevent="confirmPasswordVisible = !confirmPasswordVisible">
              <svg v-if="!confirmPasswordVisible" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" /><circle cx="12" cy="12" r="3" /></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 3l18 18M10.6 5.6c.45-.07.92-.1 1.4-.1 6 0 9.5 6.5 9.5 6.5a17.9 17.9 0 0 1-3.7 4.1M6.2 6.9A17.5 17.5 0 0 0 2.5 12S6 18.5 12 18.5c1.04 0 2.03-.2 2.95-.54M10 10a3 3 0 0 0 4 4" /></svg>
            </button>
          </label>

          <label v-if="authMode === 'register'" class="cs-auth-field">
            <span class="cs-auth-field-label">邀请码（选填）</span>
            <input v-model.trim="invitationCode" name="invitation-code" autocomplete="off" placeholder="如有邀请码，请填写" />
          </label>

          <div v-if="authMode === 'login'" class="cs-auth-forgot-row">
            <a href="/forgot" @click.prevent="switchAuthMode('forgot')">忘记密码？</a>
          </div>
          <p v-if="notice" class="cs-auth-message cs-auth-message-success" role="status">{{ notice }}</p>
          <p v-if="error" class="cs-auth-message cs-auth-message-error" role="alert">{{ error }}</p>
          <button type="submit" class="cs-auth-submit" :disabled="busy">{{ busy ? '处理中…' : authMode === 'login' ? '登录' : authMode === 'register' ? '注册' : '发送重置说明' }}</button>
        </form>
        <div class="cs-auth-bottom">
          <template v-if="authMode === 'login'">
            还没有账户？ <a href="/register" @click.prevent="switchAuthMode('register')">注册</a>
          </template>
          <template v-else>
            已有账户？ <a href="/login" @click.prevent="switchAuthMode('login')">返回登录</a>
          </template>
        </div>
      </section>
    </main>

    <div v-else-if="user" class="workspace">
      <button v-if="menuOpen" type="button" class="dashboard-mobile-backdrop" aria-label="关闭菜单" @click="menuOpen = false"></button>
      <aside class="sidebar" :class="{ 'sidebar-mobile-open': menuOpen }">
        <div class="sidebar-brand-row">
          <a href="/" class="sidebar-brand" aria-label="CodeSeam 网站首页"><img :src="logo" alt="" /><strong>CodeSeam</strong></a>
          <button type="button" class="sidebar-collapse-button" :aria-label="sidebarCollapsed ? '展开导航' : '收起导航'" @click="sidebarCollapsed = !sidebarCollapsed">
            <svg class="icon"><use href="#i-chevron-left"/></svg>
          </button>
        </div>
        <nav aria-label="用户中心导航">
          <div class="sidebar-title">我的账户</div>
          <a v-for="item in nav.filter((item) => item.group === 'account')" :key="item.id" :href="item.href" :title="item.label" class="nav-item" :class="{ selected: page === item.id }" :aria-current="page === item.id ? 'page' : undefined" @click.prevent="selectPage(item.id)">
            <svg class="icon"><use :href="'#i-' + item.icon"/></svg>
            <span>{{ item.label }}</span>

          </a>
          <div class="sidebar-title sidebar-title-secondary">获取帮助</div>
          <a v-for="item in nav.filter((item) => item.group === 'support')" :key="item.id" :href="item.href" :title="item.label" class="nav-item" :class="{ selected: page === item.id }" :aria-current="page === item.id ? 'page' : undefined" @click.prevent="selectPage(item.id)">
            <svg class="icon"><use :href="'#i-' + item.icon"/></svg><span>{{ item.label }}</span>
          </a>
        </nav>
        <div class="sidebar-spacer"></div>
        <div class="sidebar-footer">
          <a href="/#workflow" class="sidebar-footer-doc"><svg class="icon"><use href="#i-help"/></svg><span>使用文档</span><span class="sidebar-external">↗</span></a>
          <button type="button" class="logout-button" @click="signOut"><svg class="icon"><use href="#i-log-out"/></svg><span>退出登录</span></button>
        </div>
      </aside>

      <main class="content">
        <div v-if="page !== 'overview'" class="page-head">
          <div><p class="eyebrow">{{ page === 'service' ? 'SERVICE OVERVIEW' : page === 'usage' ? 'USAGE ANALYTICS' : 'CODESEAM PORTAL' }}</p>
            <h1>{{ nav.find((item) => item.id === page)?.label || '概览' }}</h1>
            <p class="subtle">{{ page === 'service' ? '了解你的服务状态、有效期限和可用额度。' : page === 'usage' ? '查看真实的 API 请求记录及 Token 使用情况。' : page === 'devices' || page === 'history' ? '这里将展示与你的连接有关的记录。' : '安装客户端并完成 Claude Code 连接。' }}</p>
          </div>
          <button v-if="page !== 'help' && page !== 'devices' && page !== 'history'" type="button" class="button button-secondary" :disabled="loading" @click="refreshData"><svg class="icon"><use href="#i-refresh"/></svg>{{ loading ? '刷新中' : '刷新数据' }}</button>
        </div>
        <p v-if="error" class="banner banner-danger page-banner" role="alert">{{ error }}</p>

        <template v-if="page === 'overview'">
          <section class="overview-welcome">
            <h1>欢迎回来，{{ userName }}</h1>
            <p>在这里管理你的 CodeSeam 服务、查看使用记录，并下载桌面客户端。</p>
            <div class="overview-welcome-account">当前账户：<span>{{ userName }}</span></div>
          </section>

          <div v-if="!primaryService || !serviceActive" class="overview-alert" role="status">
            <svg class="icon"><use href="#i-help"/></svg>
            <span>{{ primaryService ? '当前服务状态：' + stateLabel(primaryService.state) + '。请查看服务详情，了解有效期与可用额度。' : '你还未开通专属服务。开通后即可通过 CodeSeam 客户端使用完整连接服务。' }}</span>
            <a href="/app/service" @click.prevent="selectPage('service')">查看服务 →</a>
          </div>

          <section class="overview-dashboard-panel" aria-label="使用统计">
            <div class="overview-panel-tabs">
              <div class="overview-tabs">
                <button type="button" class="active" aria-current="page">概览</button>
                <button type="button" @click="selectPage('usage')">用量明细</button>
              </div>
              <button type="button" class="overview-refresh" :disabled="loading || usageLoading || activityLoading" aria-label="刷新统计" title="刷新统计" @click="refreshOverview">
                <svg class="icon"><use href="#i-refresh"/></svg>
              </button>
            </div>
            <div class="overview-panel-body">
              <div class="overview-panel-filters">
                <span class="overview-range-note">统计范围：{{ startDate }} — {{ endDate }}</span>
                <div class="overview-range-options" aria-label="统计时间范围">
                  <button type="button" :class="{ active: overviewRangeDays === 365 }" :aria-pressed="overviewRangeDays === 365" :disabled="usageLoading" @click="setRange(365)">近一年</button>
                  <button type="button" :class="{ active: overviewRangeDays === 30 }" :aria-pressed="overviewRangeDays === 30" :disabled="usageLoading" @click="setRange(30)">30 天</button>
                  <button type="button" :class="{ active: overviewRangeDays === 7 }" :aria-pressed="overviewRangeDays === 7" :disabled="usageLoading" @click="setRange(7)">7 天</button>
                </div>
              </div>
              <div class="overview-stat-grid">
                <div class="overview-stat-tile"><span>请求次数</span><strong>{{ stat(stats?.total_requests) }}</strong></div>
                <div class="overview-stat-tile"><span>输入 Tokens</span><strong>{{ stat(stats?.total_input_tokens) }}</strong></div>
                <div class="overview-stat-tile"><span>输出 Tokens</span><strong>{{ stat(stats?.total_output_tokens) }}</strong></div>
                <div class="overview-stat-tile"><span>总 Tokens</span><strong>{{ stat(stats?.total_tokens) }}</strong></div>
                <div class="overview-stat-tile"><span>缓存 Tokens</span><strong>{{ stat(totalCache) }}</strong></div>
                <div class="overview-stat-tile"><span>实际扣除额度</span><strong>{{ typeof stats?.total_actual_cost === 'number' ? stats.total_actual_cost.toFixed(4) : '—' }}</strong></div>
                <div class="overview-stat-tile"><span>套餐已使用</span><strong>{{ quotaRatio === null ? '—' : quotaRatio.toFixed(1) + '%' }}</strong></div>
                <div class="overview-stat-tile"><span>剩余有效期</span><strong>{{ dayCount === null ? '—' : dayCount + ' 天' }}</strong></div>
              </div>
              <div class="overview-activity-head">
                <span>活动 · 近一年</span>
                <span v-if="activityLoading">读取中…</span>
                <span v-else-if="activityError" class="overview-activity-error">{{ activityError }}</span>
                <span v-else-if="activityLoaded">{{ activeDays }} 个活跃日</span>
              </div>
              <div class="overview-activity-scroll">
                <div class="overview-activity-grid" :aria-label="activityLoaded ? '近一年按日 Token 活动，' + activeDays + ' 个活跃日' : '近一年活动正在加载'">
                  <span v-for="cell in activityCells" :key="cell.date" class="overview-activity-cell" :class="activityLoaded ? 'level-' + cell.level : 'level-pending'" :title="activityLoaded ? cell.date + ' · ' + formatNumber(cell.amount) + ' Tokens' : cell.date + ' · 尚未获取活动记录'"></span>
                </div>
              </div>
              <div class="overview-activity-footer"><span>365 天前</span><span>今天</span></div>
            </div>
          </section>

          <section class="overview-lower-grid" aria-label="账户与服务信息">
            <article class="card overview-service-card">
              <div class="overview-section-heading">
                <h2><img :src="logo" alt="" /> 当前服务</h2>
                <a href="/app/service" @click.prevent="selectPage('service')">查看详情 →</a>
              </div>
              <div class="overview-service-name"><strong>{{ primaryService?.name || '尚未分配专属服务' }}</strong><span :class="serviceActive ? 'status-ok' : 'status-warn'">{{ primaryService ? stateLabel(primaryService.state) : '待开通' }}</span></div>
              <div class="overview-service-fields">
                <div><span>服务到期</span><strong>{{ primaryService?.expires_at ? formatDate(primaryService.expires_at) : '—' }}</strong></div>
                <div><span>剩余天数</span><strong>{{ dayCount === null ? '—' : dayCount + ' 天' }}</strong></div>
                <div><span>剩余套餐额度</span><strong>{{ stat(remainingQuota) }}</strong></div>
              </div>
              <div class="overview-service-quota"><div :style="{ width: (quotaRatio || 0) + '%' }"></div></div>
              <p>套餐额度为平台计量，并非 Claude 官方 5h / 1w 窗口余额。</p>
            </article>
            <article class="card overview-download-card">
              <div class="overview-section-heading"><h2>CodeSeam 客户端</h2><span>Windows</span></div>
              <p>安装客户端，登录并完成环境诊断，即可开始使用受管连接。</p>
              <a v-if="downloadUrl" :href="downloadUrl" target="_blank" rel="noreferrer" class="button button-primary"><svg class="icon"><use href="#i-download"/></svg>下载 Windows 客户端</a>
              <a v-else href="/app/download" class="button button-secondary" @click.prevent="selectPage('help')">查看安装说明 <svg class="icon"><use href="#i-arrow"/></svg></a>
            </article>
          </section>
        </template>

        <template v-else-if="page === 'service'">
          <template v-if="!primaryService">
            <article class="card empty-service"><span class="large-icon"><svg class="icon"><use href="#i-shield"/></svg></span><h2>尚未开通 CodeSeam 服务</h2><p>当前账号没有可展示的专属服务。开通后刷新页面，即可查看你的服务期限与使用额度。</p><a v-if="supportUrl" class="button button-primary" :href="supportUrl" target="_blank" rel="noreferrer">联系运营<svg class="icon"><use href="#i-arrow"/></svg></a><span v-else class="subtle">开通或续费请联系你的服务提供方</span></article>
          </template>
          <template v-else>
            <div class="summary-grid">
              <article class="card summary-card"><div class="summary-icon"><svg class="icon"><use href="#i-shield"/></svg></div><span class="summary-label">服务状态</span><strong :class="serviceActive ? 'status-ok' : 'status-warn'"><span class="status-dot"></span>{{ stateLabel(primaryService.state) }}</strong><small>{{ primaryService.name || 'CodeSeam 专属服务' }}</small></article>
              <article class="card summary-card"><div class="summary-icon"><svg class="icon"><use href="#i-calendar"/></svg></div><span class="summary-label">剩余有效期</span><strong class="summary-value">{{ dayCount === null ? '未提供' : dayCount + ' 天' }}</strong><small>到期 {{ formatDate(primaryService.expires_at) }}</small></article>
              <article class="card summary-card"><div class="summary-icon"><svg class="icon"><use href="#i-chart"/></svg></div><span class="summary-label">剩余套餐额度</span><strong class="summary-value">{{ formatNumber(remainingQuota) }}</strong><small>{{ quotaRatio === null ? '当前服务未提供套餐额度' : '已使用 ' + quotaRatio.toFixed(1) + '%' }}</small></article>
            </div>
            <div class="service-columns">
              <article class="card service-panel">
                <div class="card-head"><div><h2>专属服务</h2><p>当前账号已分配的服务信息</p></div><span class="state-badge" :class="serviceActive ? 'state-active' : 'state-warning'">{{ stateLabel(primaryService.state) }}</span></div>
                <div class="service-name-row"><span class="product-icon"><svg class="icon"><use href="#i-terminal"/></svg></span><div><h3>{{ primaryService.name || 'Claude Code Dedicated' }}</h3><span class="subtle">CodeSeam · Claude Code 连接服务</span></div></div>
                <div class="detail-grid"><div><span>开始日期</span><strong>{{ formatDate(primaryService.starts_at) }}</strong></div><div><span>到期日期</span><strong>{{ formatDate(primaryService.expires_at) }}</strong></div><div><span>剩余天数</span><strong>{{ dayCount === null ? '未提供' : dayCount + ' 天' }}</strong></div></div>
                <div class="quota-section"><div class="section-row"><h3>套餐额度</h3><span>{{ quotaRatio === null ? '未提供' : formatNumber(usedQuota) + ' / ' + formatNumber(totalQuota) }}</span></div><div class="progress-track"><div class="progress-fill" :style="{ width: (quotaRatio || 0) + '%' }"></div></div><p>平台计量套餐额度，不等于 Claude 官方 5h / 1w 窗口剩余额度。</p></div>
                <div class="service-footer"><svg class="icon"><use href="#i-clock"/></svg>{{ serviceActive ? '服务已配置，可通过客户端连接使用。' : '当前服务暂不可用，请核对状态或联系运营。' }}<span v-if="lastRefresh" class="last-refresh">更新于 {{ lastRefresh }}</span></div>
              </article>
              <div class="side-stack">
                <article class="card action-panel"><div class="action-top"><span class="action-icon"><svg class="icon"><use href="#i-download"/></svg></span><span class="eyebrow">CODESEAM CLIENT</span></div><h2>开始使用 Claude Code</h2><p>安装 Windows 客户端，登录后按提示完成环境检测与连接。</p><a v-if="downloadUrl" :href="downloadUrl" target="_blank" rel="noreferrer" class="button button-primary button-wide"><svg class="icon"><use href="#i-download"/></svg>下载 Windows 客户端</a><button v-else type="button" disabled class="button button-primary button-wide"><svg class="icon"><use href="#i-download"/></svg>安装包暂未发布</button><button type="button" class="inline-link" @click="selectPage('help')">查看安装指引 <svg class="icon"><use href="#i-arrow"/></svg></button></article>
                <article class="card support-panel"><div class="support-line"><span class="support-icon"><svg class="icon"><use href="#i-help"/></svg></span><h3>续费与服务支持</h3></div><p>需要续费或遇到服务异常？首期由运营人员协助处理。</p><a v-if="supportUrl" :href="supportUrl" target="_blank" rel="noreferrer" class="inline-link">联系运营 <svg class="icon"><use href="#i-arrow"/></svg></a><span v-else class="subtle">请联系你的服务提供方</span></article>
              </div>
            </div>
          </template>
        </template>

        <template v-else-if="page === 'usage'">
          <div class="usage-toolbar card">
            <div class="range-pills"><button type="button" @click="setRange(1)">今天</button><button type="button" @click="setRange(7)">近 7 天</button><button type="button" @click="setRange(30)">近 30 天</button></div>
            <div class="range-fields"><label>开始<input v-model="startDate" type="date"/></label><span>—</span><label>结束<input v-model="endDate" type="date"/></label><button type="button" class="button button-primary button-sm" :disabled="usageLoading" @click="loadUsage">{{ usageLoading ? '查询中' : '查询' }}</button></div>
          </div>
          <div class="usage-metrics">
            <article class="card metric"><span>总请求次数</span><strong>{{ formatNumber(stats?.total_requests) }}</strong><small>所选日期范围内</small></article>
            <article class="card metric"><span>总 Token</span><strong>{{ formatNumber(stats?.total_tokens) }}</strong><small>平台真实计量</small></article>
            <article class="card metric"><span>Input Token</span><strong>{{ formatNumber(stats?.total_input_tokens) }}</strong><small>输入 Token</small></article>
            <article class="card metric"><span>Output Token</span><strong>{{ formatNumber(stats?.total_output_tokens) }}</strong><small>输出 Token</small></article>
          </div>
          <article class="card trend-panel"><div class="card-head"><div><h2>每日用量趋势</h2><p>以服务端按日统计的总 Token 绘制</p></div><span class="chart-range">{{ startDate }} — {{ endDate }}</span></div>
            <div v-if="trend.length" class="chart-bars"><div v-for="point in trend" :key="point.date" class="chart-column" :title="point.date + ' · ' + formatNumber(point.total_tokens) + ' Tokens'"><span class="chart-count">{{ formatNumber(point.total_tokens) }}</span><span class="chart-rail"><span class="chart-bar" :style="{ height: barHeight(point) + 'px' }"></span></span><span class="chart-date">{{ point.date.slice(5) }}</span></div></div>
            <div v-else class="chart-empty">当前日期范围暂无用量趋势数据</div>
          </article>
          <div class="usage-bottom"><article class="card metric-detail"><h2>用量构成</h2><div class="data-row"><span>Cache Token</span><strong>{{ formatNumber(totalCache) }}</strong></div><div class="data-row"><span>实际扣除额度</span><strong>{{ typeof stats?.total_actual_cost === 'number' ? stats.total_actual_cost.toFixed(4) : '未提供' }}</strong></div><div class="data-row"><span>查询范围</span><strong>{{ startDate }} — {{ endDate }}</strong></div></article><article class="card usage-note"><span class="note-icon"><svg class="icon"><use href="#i-help"/></svg></span><h2>关于用量数据</h2><p>本页展示 Sub2API 平台计量记录，与 Claude 官方 5 小时 / 每周窗口属于不同口径。官方剩余额度尚无可靠接口，因此不会在这里估算。</p></article></div>
        </template>

        <template v-else-if="page === 'help'">
          <article class="card download-hero"><div class="download-hero-copy"><p class="eyebrow">DESKTOP CLIENT</p><h2>从 CodeSeam 客户端开始</h2><p>使用 Windows 客户端登录你的账号，由客户端完成 Claude Code 的配置检查和受管连接。</p><a v-if="downloadUrl" :href="downloadUrl" target="_blank" rel="noreferrer" class="button button-primary"><svg class="icon"><use href="#i-download"/></svg>下载 Windows 版本</a><button v-else disabled class="button button-primary"><svg class="icon"><use href="#i-download"/></svg>安装包暂未发布</button></div><div class="download-visual"><svg class="icon"><use href="#i-terminal"/></svg><span>Claude Code</span><small>CodeSeam Managed Connection</small></div></article>
          <div class="guide-title"><h2>三步开始使用</h2><span>无需手动填写服务 Key</span></div>
          <div class="guide-grid"><article class="card guide-card"><span class="step-num">01</span><h3>下载并安装</h3><p>获取正式 Windows 安装包，安装 CodeSeam 客户端。</p></article><article class="card guide-card"><span class="step-num">02</span><h3>登录与环境检测</h3><p>使用门户账号登录，按照客户端提示检查配置与网络出口。</p></article><article class="card guide-card"><span class="step-num">03</span><h3>启动 Claude Code</h3><p>确认诊断结果后，在客户端启动受管 Claude Code 会话。</p></article></div>
          <article class="card help-strip"><div><h3>遇到连接问题？</h3><p>优先在客户端运行环境诊断，检查错误详情及可修复项；未覆盖的网络路径以实际诊断结果为准。</p></div><a v-if="supportUrl" :href="supportUrl" target="_blank" rel="noreferrer" class="button button-secondary">联系运营<svg class="icon"><use href="#i-arrow"/></svg></a><span v-else class="subtle">如需人工协助，请联系服务提供方</span></article>
        </template>
        <template v-else>
          <div class="planned-page card">
            <svg class="icon"><use :href="page === 'devices' ? '#i-monitor' : '#i-clock'"/></svg>
            <h2>{{ page === 'devices' ? '设备管理' : '历史记录' }}</h2>
            <p>这一模块的界面入口已经保留，实际数据和交互将在对应服务能力接入后开放。</p>
            <span>尚未开放 · 暂无数据</span>
          </div>
        </template>
        <footer class="content-footer">CodeSeam · 用户服务门户 <span>服务状态与用量均以实际服务端记录为准</span></footer>
      </main>
    </div>
  </div>
</template>
