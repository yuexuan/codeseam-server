import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'

const api = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, ''),
  withCredentials: true,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

export type User = { id: number; email?: string; username?: string; display_name?: string; status?: string }
export type AuthResponse = { access_token: string; refresh_token?: string; expires_in?: number; user?: User }
export type Service = { id: number; name: string; state: string; starts_at?: string | null; expires_at?: string | null; quota?: number; quota_used?: number }
export type UsageStats = { total_requests?: number; total_input_tokens?: number; total_output_tokens?: number; total_cache_tokens?: number; total_cache_read_tokens?: number; total_cache_creation_tokens?: number; total_tokens?: number; total_actual_cost?: number; total_cost?: number }
export type TrendPoint = { date: string; requests?: number; total_tokens?: number; input_tokens?: number; output_tokens?: number; cache_creation_tokens?: number; cache_read_tokens?: number }

let refreshing: Promise<string> | null = null
function saveAuth(data: AuthResponse) {
  localStorage.setItem('auth_token', data.access_token)
  if (data.refresh_token) localStorage.setItem('refresh_token', data.refresh_token)
  if (data.expires_in) localStorage.setItem('token_expires_at', String(Date.now() + data.expires_in * 1000))
  if (data.user) localStorage.setItem('auth_user', JSON.stringify(data.user))
}
export function clearAuth() { ['auth_token', 'refresh_token', 'token_expires_at', 'auth_user'].forEach((key) => localStorage.removeItem(key)) }
export function hasAuth() { return Boolean(localStorage.getItem('auth_token') || localStorage.getItem('refresh_token')) }

function unwrap<T>(value: unknown): T {
  const body = value as { code?: number; data?: T; message?: string }
  if (body && typeof body === 'object' && 'code' in body) {
    if (body.code !== 0) throw new Error(body.message || '请求失败')
    return body.data as T
  }
  return value as T
}

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  config.headers['Accept-Language'] = 'zh-CN'
  return config
})
api.interceptors.response.use((response) => { response.data = unwrap(response.data); return response }, async (error: AxiosError) => {
  const config = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined
  if (error.response?.status === 401 && config && !config._retry && localStorage.getItem('refresh_token') && !['/auth/login', '/auth/register', '/auth/refresh', '/auth/logout'].some((path) => String(config.url).includes(path))) {
    config._retry = true
    refreshing ||= (async () => {
      const refreshToken = localStorage.getItem('refresh_token')
      const result = await axios.post(`${api.defaults.baseURL}/auth/refresh`, { refresh_token: refreshToken }, { withCredentials: true })
      const data = unwrap<AuthResponse>(result.data)
      saveAuth(data)
      return data.access_token
    })().finally(() => { refreshing = null })
    try { config.headers.Authorization = `Bearer ${await refreshing}`; return api(config) } catch { clearAuth() }
  }
  return Promise.reject(error)
})

export async function login(email: string, password: string) { const { data } = await api.post<AuthResponse>('/auth/login', { email, password }); saveAuth(data); return data }
export async function register(email: string, password: string, invitationCode = '') { const { data } = await api.post<AuthResponse>('/auth/register', { email, password, ...(invitationCode ? { invitation_code: invitationCode } : {}) }); saveAuth(data); return data }
export async function forgotPassword(email: string) { const { data } = await api.post<{ message?: string }>('/auth/forgot-password', { email }); return data }
export async function logout() { const token = localStorage.getItem('refresh_token'); try { if (token) await api.post('/auth/logout', { refresh_token: token }) } finally { clearAuth() } }
export async function me() { const { data } = await api.get<User>('/auth/me'); return data }
export async function currentService() { const { data } = await api.get<{ services?: Service[] }>('/codeseam/service/current'); return data }
export async function usageStats(startDate: string, endDate: string) { const { data } = await api.get<UsageStats>('/usage/stats', { params: { start_date: startDate, end_date: endDate } }); return data }
export async function usageTrend(startDate: string, endDate: string) { const { data } = await api.get<{ trend?: TrendPoint[] }>('/usage/dashboard/trend', { params: { start_date: startDate, end_date: endDate, granularity: 'day' } }); return data }
export default api
