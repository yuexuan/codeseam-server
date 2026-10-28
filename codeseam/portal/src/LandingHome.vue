<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import logo from './assets/codeseam-logo.svg'
import LandingIcon from './LandingIcon.vue'

const menuOpen = ref(false)
const openFaq = ref<number | null>(0)
const theme = ref<'light' | 'dark'>(
  localStorage.getItem('cs.portal.theme') === 'dark' ? 'dark' : 'light'
)
const downloadUrl = import.meta.env.VITE_WINDOWS_DOWNLOAD_URL || ''
const supportUrl = import.meta.env.VITE_SUPPORT_URL || ''

const painPoints = [
  { icon: 'switch', title: '频繁切换，打断工作节奏', body: '账号、连接配置和开发环境来回调整，让本该专注的编码时间消耗在维护上。', answer: '把常见配置操作收拢到客户端，减少重复处理。' },
  { icon: 'network', title: '网络出口不一致', body: '本地请求和服务端请求走不同线路，出现问题时难以发现和定位。', answer: '集中管理连接方式，提供出口及环境诊断。' },
  { icon: 'activity', title: '状态不透明', body: '服务是否正常、何时到期、已经使用多少，都需要有明确的查看入口。', answer: '在用户中心集中查看服务期限与用量记录。' },
  { icon: 'sliders', title: '复杂的安装与配置', body: '从首次安装到启动 Claude Code，手动改动越多，越难维护。', answer: '通过 CodeSeam 客户端完成初始化和日常连接。' }
]
const featureCards = [
  { icon: 'link', tag: 'CONNECTION', title: '专属连接', body: '面向独享服务的连接与资源绑定，统一管理客户端和服务端的访问路径。' },
  { icon: 'radar', tag: 'DIAGNOSTICS', title: '环境诊断', body: '检查代理、出口 IP、DNS、IPv6 和配置变化，让问题发生的位置更清晰。' },
  { icon: 'chart', tag: 'USAGE', title: '用量可见', body: '在用户门户查看使用记录、Token 统计、套餐额度和有效期限。' },
  { icon: 'monitor', tag: 'WORKFLOW', title: '桌面优先', body: '围绕 Claude Code 的真实工作流设计，减少反复修改 CLI 配置的操作。' },
  { icon: 'shield', tag: 'PROTECTION', title: '连接保护', body: '受管代理不可用时拒绝无意降级直连，配合运行状态与故障提示。' },
  { icon: 'refresh', tag: 'RECOVERY', title: '配置可恢复', body: '提供配置备份、恢复与重新连接入口，减少切换与故障后的手工处理。' }
]
const faqs = [
  { q: 'CodeSeam 是什么？', a: 'CodeSeam 是围绕 Claude Code 构建的桌面连接客户端与用户服务门户，重点是连接管理、环境诊断、专属服务以及用量查询。' },
  { q: '如何开始使用？', a: '先创建并登录 CodeSeam 账号。服务开通后安装 Windows 客户端，按提示检查环境并连接。具体安装版本以下载页实际发布内容为准。' },
  { q: '可以查看哪些用量数据？', a: '用户门户支持按日期查看请求统计及 Token 使用趋势，也可以查看服务状态、有效期和平台套餐额度；官方 5 小时及 1 周窗口需以实际接入能力为准。' },
  { q: 'CodeSeam 会自动处理所有网络问题吗？', a: '不会。客户端可以诊断部分网络与配置问题，并防止特定情况下意外直连；实际连接质量仍取决于所选线路与上游服务。' },
  { q: '支持哪些系统？', a: '当前以 Windows 客户端为核心交付对象。其他平台的支持状态将在正式发布后明确列出。' },
  { q: '套餐价格与售后如何确认？', a: '当前首页展示服务方案的结构，正式价格、服务范围与售后条款将在产品上线前单独发布，不以此页面的框架示意作为购买承诺。' }
]

watch(theme, (current) => {
  document.documentElement.dataset.theme = current
  localStorage.setItem('cs.portal.theme', current)
}, { immediate: true })
onMounted(() => { document.title = 'CodeSeam · 专注连接，安心编码' })
function toggleTheme() { theme.value = theme.value === 'light' ? 'dark' : 'light' }
function closeMenu() { menuOpen.value = false }
function toggleFaq(index: number) { openFaq.value = openFaq.value === index ? null : index }
</script>

<template>
  <div class="landing">
    <header class="landing-header">
      <div class="landing-container landing-nav">
        <a href="#home" class="landing-brand" aria-label="CodeSeam 首页" @click="closeMenu">
          <img :src="logo" alt="" /> <span>CodeSeam</span>
        </a>
        <nav class="landing-nav-links" :class="{ 'landing-nav-open': menuOpen }" aria-label="首页导航">
          <a href="#advantages" @click="closeMenu">产品优势</a>
          <a href="#workflow" @click="closeMenu">工作流程</a>
          <a href="#features" @click="closeMenu">核心能力</a>
          <a href="#plans" @click="closeMenu">服务方案</a>
          <a href="#faq" @click="closeMenu">常见问题</a>
          <a href="/app" class="landing-nav-mobile-account" @click="closeMenu">登录用户中心</a>
        </nav>
        <div class="landing-nav-actions">
          <button type="button" class="landing-theme" :aria-label="theme === 'light' ? '切换深色模式' : '切换浅色模式'" @click="toggleTheme">
            <svg v-if="theme === 'light'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="M20 15.5A8.5 8.5 0 0 1 8.5 4 8.5 8.5 0 1 0 20 15.5Z" /></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M4.93 4.93l1.42 1.42m11.3 11.3 1.42 1.42M2 12h2m16 0h2M4.93 19.07l1.42-1.42m11.3-11.3 1.42-1.42" /></svg>
          </button>
          <a href="/login" class="landing-signin">登录</a>
          <a href="/app" class="landing-btn landing-btn-primary landing-nav-cta">开始使用 <span aria-hidden="true">↗</span></a>
          <button type="button" class="landing-menu-button" :aria-expanded="menuOpen" aria-label="切换导航菜单" @click="menuOpen = !menuOpen">
            <span></span><span></span><span></span>
          </button>
        </div>
      </div>
    </header>

    <main id="home">
      <section class="landing-hero landing-section">
        
        <div class="landing-container landing-hero-grid">
          <div class="landing-hero-copy">
            <div class="landing-kicker"><span class="landing-kicker-dot"></span> 专为 Claude Code 打造的连接工具</div>
            <h1>让连接更简单。<br /><span>让编码更专注。</span></h1>
            <p class="landing-hero-lead">CodeSeam 将受管连接、环境诊断和服务管理集中在一个简单的桌面工具里。少一些反复配置，多一些连贯的开发时间。</p>
            <div class="landing-hero-actions">
              <a class="landing-btn landing-btn-primary landing-btn-lg" href="/app">进入用户中心 <span aria-hidden="true">↗</span></a>
              <a class="landing-btn landing-btn-outline landing-btn-lg" href="#features">了解产品能力 <span aria-hidden="true">↓</span></a>
            </div>
            <div class="landing-hero-notes">
              <span>Windows 桌面客户端</span>
              <span>专属服务管理</span>
              <span>用量记录可查</span>
            </div>
          </div>

          <div class="landing-hero-art" aria-label="CodeSeam 桌面客户端界面示意">
            
            
            
            <div class="landing-desktop">
              <div class="landing-desktop-bar"><span class="landing-desktop-identity"><img :src="logo" alt="" /> CodeSeam</span><span class="landing-desktop-window">— &nbsp; □ &nbsp; ×</span></div>
              <div class="landing-desktop-body">
                <aside class="landing-desktop-side">
                  <div class="landing-mini-nav selected"><LandingIcon name="monitor" /><span>总览</span></div>
                  <div class="landing-mini-nav"><LandingIcon name="radar" /><span>环境诊断</span></div>
                  <div class="landing-mini-nav"><LandingIcon name="terminal" /><span>Claude Code</span></div>
                  <div class="landing-mini-nav"><LandingIcon name="chart" /><span>用量管理</span></div>
                  <div class="landing-mini-nav"><LandingIcon name="sliders" /><span>设置</span></div>
                </aside>
                <div class="landing-desktop-main">
                  <div class="landing-desk-label">WORKSPACE OVERVIEW</div>
                  <strong class="landing-desk-title">一切准备就绪</strong>
                  <p class="landing-desk-subtitle">你的连接状态，一目了然。</p>
                  <div class="landing-desk-status">
                    <div class="landing-desk-icon"><LandingIcon name="link" /></div>
                    <div><small>服务状态 · 示例</small><strong>受管连接</strong></div>
                    <span class="landing-desk-badge">● 正常</span>
                  </div>
                  <div class="landing-desk-split">
                    <div><small>服务有效期</small><strong>在用户中心查看</strong></div>
                    <div><small>出口检测</small><strong>启动时检查</strong></div>
                  </div>
                  <div class="landing-desk-bottom"><span>连接与诊断</span><span>↗</span></div>
                </div>
              </div>
            </div>
            
            <div class="landing-art-note">CodeSeam 客户端界面示意</div>
          </div>
        </div>
      </section>

      <section class="landing-value-strip" aria-label="产品定位">
        <div class="landing-container landing-value-inner">
          <div><span class="landing-value-symbol" aria-hidden="true"><LandingIcon name="link" /></span><strong>更可控的连接</strong><small>统一管理服务与配置</small></div>
          <div><span class="landing-value-symbol" aria-hidden="true"><LandingIcon name="radar" /></span><strong>更清楚的诊断</strong><small>从异常到原因都有线索</small></div>
          <div><span class="landing-value-symbol" aria-hidden="true"><LandingIcon name="chart" /></span><strong>更直观的用量</strong><small>数据与有效期集中查看</small></div>
        </div>
      </section>

      <section id="advantages" class="landing-section landing-advantages">
        <div class="landing-container">
          <div class="landing-section-head landing-head-split">
            <div><h2>我们解决的，<br />是编码之外的麻烦。</h2></div>
            <p>CodeSeam 不改变你的编程方式。它把连接、诊断与服务管理统一起来，让你把时间留给真正重要的开发工作。</p>
          </div>
          <div class="landing-pain-grid">
            <article v-for="item in painPoints" :key="item.title" class="landing-pain-card">
              <span class="landing-pain-icon"><LandingIcon :name="item.icon" /></span>
              <h3>{{ item.title }}</h3>
              <p>{{ item.body }}</p>
              <div class="landing-pain-answer"><span class="landing-answer-check" aria-hidden="true">↗</span>{{ item.answer }}</div>
            </article>
          </div>
          <div class="landing-promise">
            
            <div><strong>专注连接，不打断创作。</strong><p>从启动前检查，到连接管理，再到用量查看，围绕一个连续的开发过程展开。</p></div>
            <a href="#features">探索核心能力 <span aria-hidden="true">↗</span></a>
          </div>
        </div>
      </section>

      <section id="workflow" class="landing-section landing-workflow">
        <div class="landing-container">
          <div class="landing-section-head landing-center"><h2>准备好后，直接开始。</h2><p>从用户中心到桌面客户端，自然衔接你的 Claude Code 工作环境。</p></div>
          <div class="landing-steps">
            <article><div class="landing-step-icon"><LandingIcon name="user" /></div><h3>登录并开通服务</h3><p>在用户中心创建账号，查看已开通的专属服务与有效期。</p><a href="/app">进入用户中心 ↗</a></article>
            <article><div class="landing-step-icon"><LandingIcon name="download" /></div><h3>安装 CodeSeam</h3><p>下载客户端，按界面提示检查代理、出口和本机环境。</p><a v-if="downloadUrl" :href="downloadUrl" target="_blank" rel="noreferrer">下载 Windows 版 ↗</a><span v-else class="landing-step-unavailable">客户端下载即将开放</span></article>
            <article><div class="landing-step-icon"><LandingIcon name="terminal" /></div><h3>启动 Claude Code</h3><p>完成受管配置后，从客户端进入你的日常编码工作流。</p><a href="#features">查看支持的能力 ↗</a></article>
          </div>
        </div>
      </section>

      <section id="features" class="landing-section landing-features">
        <div class="landing-container">
          <div class="landing-section-head landing-center"><h2>功能清楚，连接从容。</h2><p>围绕桌面客户端和用户门户，建立完整、清晰的使用体验。</p></div>
          <div class="landing-feature-grid">
            <article v-for="feature in featureCards" :key="feature.tag" class="landing-feature-card">
              <div class="landing-feature-icon"><LandingIcon :name="feature.icon" /></div>
              <h3>{{ feature.title }}</h3><p>{{ feature.body }}</p>
            </article>
          </div>
          <div class="landing-feature-banner">
            <div><span class="landing-feature-banner-kicker">CODESEAM DESKTOP + PORTAL</span><h3>桌面连接，与网页服务中心，彼此衔接。</h3><p>桌面端负责日常连接与环境检查；网页端负责账号、服务期限、使用记录与帮助入口。</p></div>
            <a class="landing-btn landing-btn-light" href="/app">打开用户中心 ↗</a>
          </div>
        </div>
      </section>

      <section id="plans" class="landing-section landing-plans">
        <div class="landing-container">
          <div class="landing-section-head landing-center"><h2>适合自己的，才是好的选择。</h2><p>首期聚焦专属服务。以下是套餐展示框架，具体售价与权益将在发布时确认。</p></div>
          <div class="landing-plan-grid">
            <article class="landing-plan-card"><p class="landing-plan-label">DEDICATED · STANDARD</p><h3>专属 5x</h3><p class="landing-plan-desc">适合稳定、持续的日常开发需求。</p><div class="landing-plan-price">即将公布 <span>/ 月</span></div><div class="landing-plan-line"></div><ul><li>专属服务配置</li><li>桌面客户端连接</li><li>环境诊断与用量查询</li></ul><a href="/app" class="landing-btn landing-btn-outline landing-plan-action">了解服务 ↗</a></article>
            <article class="landing-plan-card landing-plan-featured"><div class="landing-plan-highlight">专属方案</div><p class="landing-plan-label">DEDICATED · ADVANCED</p><h3>专属 20x</h3><p class="landing-plan-desc">为更高强度的开发工作预留选择。</p><div class="landing-plan-price">即将公布 <span>/ 月</span></div><div class="landing-plan-line"></div><ul><li>专属服务配置</li><li>桌面客户端连接</li><li>环境诊断与用量查询</li></ul><a href="/app" class="landing-btn landing-btn-primary landing-plan-action">了解服务 ↗</a></article>
            <article class="landing-plan-card"><p class="landing-plan-label">NEED SOME HELP?</p><h3>购买前咨询</h3><p class="landing-plan-desc">有连接方式、方案选择和部署相关疑问？</p><div class="landing-plan-price landing-plan-consult">先聊聊需求</div><div class="landing-plan-line"></div><ul><li>了解交付形式</li><li>确认平台与系统要求</li><li>查看当前上线状态</li></ul><a v-if="supportUrl" :href="supportUrl" target="_blank" rel="noreferrer" class="landing-btn landing-btn-outline landing-plan-action">联系我们 ↗</a><a v-else href="#faq" class="landing-btn landing-btn-outline landing-plan-action">查看常见问题 ↗</a></article>
          </div>
          <p class="landing-plan-note">* 页面仅用于产品结构设计，不代表套餐已经开放购买或提供确定的上游订阅额度。</p>
        </div>
      </section>

      <section id="faq" class="landing-section landing-faq">
        <div class="landing-container landing-faq-layout">
          <div class="landing-faq-intro"><h2>你可能还想了解。</h2><p>关于 CodeSeam、客户端和服务的常见问题，后续会随着实际交付持续完善。</p><a class="landing-text-link" href="/app">前往用户中心 <span>↗</span></a></div>
          <div class="landing-faq-list">
            <div v-for="(item, index) in faqs" :key="item.q" class="landing-faq-item" :class="{ 'is-open': openFaq === index }">
              <button type="button" :aria-expanded="openFaq === index" @click="toggleFaq(index)"><span>{{ item.q }}</span><span class="landing-faq-symbol">{{ openFaq === index ? '−' : '+' }}</span></button>
              <p v-if="openFaq === index">{{ item.a }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="landing-bottom-cta">
        <div class="landing-container landing-bottom-inner">
          <h2>专注你的下一行代码。</h2>
          <p>把繁琐留给连接工具，把创造留给自己。</p>
          <a href="/app" class="landing-btn landing-btn-light landing-btn-lg">进入 CodeSeam <span aria-hidden="true">↗</span></a>
        </div>
      </section>
    </main>

    <footer class="landing-footer">
      <div class="landing-container landing-footer-grid">
        <div><a href="#home" class="landing-brand landing-footer-brand"><img :src="logo" alt="" /><span>CodeSeam</span></a><p>专注连接，安心编码。<br />A more seamless coding experience.</p></div>
        <div><h4>产品</h4><a href="#advantages">产品优势</a><a href="#features">核心能力</a><a href="#workflow">开始使用</a></div>
        <div><h4>服务</h4><a href="#plans">服务方案</a><a href="/app">用户中心</a><a href="#faq">常见问题</a></div>
        <div><h4>获取帮助</h4><a href="#faq">使用说明</a><a v-if="supportUrl" :href="supportUrl" target="_blank" rel="noreferrer">联系支持</a><span v-else>支持渠道即将公布</span></div>
      </div>
      <div class="landing-container landing-footer-bottom"><span>© {{ new Date().getFullYear() }} CodeSeam. All rights reserved.</span><span>Built for the way you code.</span></div>
    </footer>
  </div>
</template>
