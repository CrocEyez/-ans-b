<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:23456'

const activePanel = ref('review')

const knowledgeForm = reactive({
  question: '',
  answer: '',
  category: '',
  tags: '',
  source: '',
  remark: '',
})

const askForm = reactive({
  question: '',
})

const adminForm = reactive({
  username: '',
  password: '',
})

const reviewForm = reactive({
  question: '',
  answer: '',
  category: '',
  tags: '',
  source: '',
  remark: '',
  reviewerNote: '',
})

const submitLoading = ref(false)
const askLoading = ref(false)
const adminLoading = ref(false)
const submissionsLoading = ref(false)
const approveLoading = ref(false)
const rejectLoading = ref(false)
const enterReviewLoading = ref(false)
const submitMessage = ref('')
const submitError = ref('')
const askError = ref('')
const askResult = ref(null)
const adminMessage = ref('')
const adminError = ref('')
const reviewMessage = ref('')
const reviewError = ref('')
const adminToken = ref(window.localStorage.getItem('ans-b-admin-token') || '')
const adminUser = ref(JSON.parse(window.localStorage.getItem('ans-b-admin-user') || 'null'))
const reviewStatus = ref('pending')
const submissions = ref([])
const selectedSubmission = ref(null)
const reviewEntered = ref(false)

const knowledgeItems = ref([])
const knowledgeLoading = ref(false)
const knowledgeError = ref('')
const knowledgePage = ref(1)
const knowledgeTotal = ref(0)
const knowledgePageSize = 10
const knowledgeJumpPage = ref(null)

const users = ref([])
const userLoading = ref(false)
const userError = ref('')
const userPage = ref(1)
const userTotal = ref(0)
const userPageSize = 10
const userJumpPage = ref(null)

const statusOptions = [
  { value: 'pending', label: '待审核', emptyText: '暂无待审核投稿' },
  { value: 'approved', label: '已通过', emptyText: '暂无已通过投稿' },
  { value: 'rejected', label: '已驳回', emptyText: '暂无已驳回投稿' },
  { value: '', label: '全部', emptyText: '暂无投稿' },
]

const statusThemeMap = {
  pending: 'warning',
  approved: 'success',
  rejected: 'danger',
}

const navItems = [
  { key: 'review', label: '审核面板' },
  { key: 'entry', label: '知识录入' },
  { key: 'qa', label: '问答测试' },
  { key: 'knowledge', label: '知识库浏览' },
  { key: 'users', label: '用户管理' },
]

const canSubmitKnowledge = computed(() => (
  knowledgeForm.question.trim() &&
  knowledgeForm.answer.trim() &&
  !submitLoading.value
))

const canAsk = computed(() => askForm.question.trim() && !askLoading.value && isAdminLoggedIn.value)
const isAdminLoggedIn = computed(() => Boolean(adminToken.value))
const canLoginAdmin = computed(() => (
  adminForm.username.trim() &&
  adminForm.password.trim() &&
  !adminLoading.value
))
const canEnterReview = computed(() => isAdminLoggedIn.value && !enterReviewLoading.value)
const isReviewBusy = computed(() => approveLoading.value || rejectLoading.value)
const isSelectedPending = computed(() => selectedSubmission.value?.status === 'pending')
const canEditReviewForm = computed(() => isSelectedPending.value && !isReviewBusy.value)
const canApproveSubmission = computed(() => (
  selectedSubmission.value &&
  isSelectedPending.value &&
  reviewForm.question.trim() &&
  reviewForm.answer.trim() &&
  !isReviewBusy.value
))
const canRejectSubmission = computed(() => (
  selectedSubmission.value &&
  isSelectedPending.value &&
  !isReviewBusy.value
))
const emptySubmissionText = computed(() => (
  statusOptions.find((item) => item.value === reviewStatus.value)?.emptyText || '暂无投稿'
))

const totalKnowledgePages = computed(() =>
  Math.max(1, Math.ceil(knowledgeTotal.value / knowledgePageSize))
)

const totalUserPages = computed(() =>
  Math.max(1, Math.ceil(userTotal.value / userPageSize))
)

function parseTags(value) {
  return value
    .split(/[,，\n]/)
    .map((tag) => tag.trim())
    .filter(Boolean)
}

function statusLabel(status) {
  return statusOptions.find((item) => item.value === status)?.label || status || '未知'
}

function statusTheme(status) {
  return statusThemeMap[status] || 'default'
}

function formatDateTime(value) {
  if (!value) return '-'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'

  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function isAdminAuthError(error) {
  const message = String(error?.message || '')
  return message.includes('login session') ||
    message.includes('missing current user') ||
    message.includes('permission denied') ||
    message.includes('HTTP 401')
}

function clearAdminSession() {
  adminToken.value = ''
  adminUser.value = null
  submissions.value = []
  selectedSubmission.value = null
  reviewEntered.value = false
  window.localStorage.removeItem('ans-b-admin-token')
  window.localStorage.removeItem('ans-b-admin-user')
}

function handleAdminAuthError(error) {
  if (!isAdminAuthError(error)) return false
  clearAdminSession()
  adminError.value = '管理员登录已失效，请重新登录'
  return true
}

async function requestJSON(path, options = {}) {
  const {
    method = 'GET',
    payload,
    auth = false,
  } = options
  const headers = {}
  if (payload !== undefined) {
    headers['Content-Type'] = 'application/json'
  }
  if (auth && adminToken.value) {
    headers.Authorization = `Bearer ${adminToken.value}`
  }
  const response = await fetch(`${apiBaseURL}${path}`, {
    method,
    headers,
    body: payload === undefined ? undefined : JSON.stringify(payload),
  })
  const result = await response.json().catch(() => null)
  if (!response.ok || result?.code !== 0) {
    throw new Error(result?.message || `HTTP ${response.status}`)
  }
  return result.data
}

async function postJSON(path, payload) {
  return requestJSON(path, { method: 'POST', payload })
}

async function loginAdmin() {
  if (!canLoginAdmin.value) return

  adminLoading.value = true
  adminMessage.value = ''
  adminError.value = ''

  try {
    const result = await requestJSON('/api/v1/auth/admin/login', {
      method: 'POST',
      payload: {
        username: adminForm.username.trim(),
        password: adminForm.password.trim(),
      },
    })
    adminToken.value = result.token
    adminUser.value = result.user
    window.localStorage.setItem('ans-b-admin-token', result.token)
    window.localStorage.setItem('ans-b-admin-user', JSON.stringify(result.user))
    adminForm.password = ''
    reviewEntered.value = false
    submissions.value = []
    selectedSubmission.value = null
    adminMessage.value = `已登录：${result.user?.username || adminForm.username.trim()}，请进入审核端`
  } catch (error) {
    adminError.value = error.message
  } finally {
    adminLoading.value = false
  }
}

function logoutAdmin() {
  clearAdminSession()
  adminMessage.value = ''
  adminError.value = ''
}

async function enterReviewPanel() {
  if (!canEnterReview.value) return

  enterReviewLoading.value = true
  adminError.value = ''
  reviewError.value = ''
  reviewMessage.value = ''

  try {
    const loaded = await loadSubmissions({ force: true, preserveFeedback: true })
    if (loaded) {
      reviewEntered.value = true
      adminMessage.value = ''
    } else if (!adminError.value) {
      adminError.value = reviewError.value || '进入审核端失败'
    }
  } finally {
    enterReviewLoading.value = false
  }
}

function fillReviewForm(submission, options = {}) {
  const { clearFeedback = true, allowBusy = false } = options
  if (isReviewBusy.value && !allowBusy) return
  selectedSubmission.value = submission
  reviewForm.question = submission?.question || ''
  reviewForm.answer = submission?.answer || ''
  reviewForm.category = submission?.category || ''
  reviewForm.tags = Array.isArray(submission?.tags) ? submission.tags.join('，') : ''
  reviewForm.source = submission?.source || ''
  reviewForm.remark = submission?.remark || ''
  reviewForm.reviewerNote = submission?.reviewer_note || ''
  if (clearFeedback) {
    reviewMessage.value = ''
    reviewError.value = ''
  }
}

async function loadSubmissions(options = {}) {
  if (!isAdminLoggedIn.value) return false
  if (isReviewBusy.value && !options.force) return false

  const { preserveFeedback = false } = options
  submissionsLoading.value = true
  if (!preserveFeedback) {
    reviewError.value = ''
  }

  try {
    const query = reviewStatus.value ? `?status=${encodeURIComponent(reviewStatus.value)}` : ''
    const result = await requestJSON(`/api/v1/submissions${query}`, { auth: true })
    submissions.value = result
    const selected = result.find((item) => item.id === selectedSubmission.value?.id)
    fillReviewForm(selected || result[0] || null, {
      allowBusy: options.force,
      clearFeedback: !preserveFeedback,
    })
    return true
  } catch (error) {
    if (!preserveFeedback) {
      reviewMessage.value = ''
    }
    if (!handleAdminAuthError(error)) {
      reviewError.value = error.message
    }
    return false
  } finally {
    submissionsLoading.value = false
  }
}

async function approveSubmission() {
  if (!canApproveSubmission.value) return

  approveLoading.value = true
  reviewMessage.value = ''
  reviewError.value = ''

  try {
    await requestJSON(`/api/v1/submissions/${selectedSubmission.value.id}/approve`, {
      method: 'POST',
      auth: true,
      payload: {
        question: reviewForm.question.trim(),
        answer: reviewForm.answer.trim(),
        category: reviewForm.category.trim(),
        tags: parseTags(reviewForm.tags),
        source: reviewForm.source.trim(),
        remark: reviewForm.remark.trim(),
        reviewer_note: reviewForm.reviewerNote.trim(),
      },
    })
    reviewMessage.value = '审核通过，已生成向量并进入知识库'
    await loadSubmissions({ preserveFeedback: true, force: true })
  } catch (error) {
    if (!handleAdminAuthError(error)) {
      reviewError.value = error.message
    }
  } finally {
    approveLoading.value = false
  }
}

async function rejectSubmission() {
  if (!canRejectSubmission.value) return

  rejectLoading.value = true
  reviewMessage.value = ''
  reviewError.value = ''

  try {
    await requestJSON(`/api/v1/submissions/${selectedSubmission.value.id}/reject`, {
      method: 'POST',
      auth: true,
      payload: {
        reviewer_note: reviewForm.reviewerNote.trim(),
      },
    })
    reviewMessage.value = '已驳回投稿'
    await loadSubmissions({ preserveFeedback: true, force: true })
  } catch (error) {
    if (!handleAdminAuthError(error)) {
      reviewError.value = error.message
    }
  } finally {
    rejectLoading.value = false
  }
}

async function submitKnowledge() {
  if (!canSubmitKnowledge.value) return

  submitLoading.value = true
  submitMessage.value = ''
  submitError.value = ''

  try {
    await postJSON('/api/v1/knowledge', {
      question: knowledgeForm.question.trim(),
      answer: knowledgeForm.answer.trim(),
      category: knowledgeForm.category.trim(),
      tags: parseTags(knowledgeForm.tags),
      source: knowledgeForm.source.trim(),
      remark: knowledgeForm.remark.trim(),
    })
    submitMessage.value = '知识已写入数据库并完成向量化'
    knowledgeForm.question = ''
    knowledgeForm.answer = ''
    knowledgeForm.tags = ''
    knowledgeForm.remark = ''
  } catch (error) {
    submitError.value = error.message
  } finally {
    submitLoading.value = false
  }
}

async function askQuestion() {
  if (!canAsk.value) return

  askLoading.value = true
  askError.value = ''
  askResult.value = null

  try {
    askResult.value = await requestJSON('/api/v1/admin/qa/ask', {
      method: 'POST',
      payload: {
        question: askForm.question.trim(),
        limit: 5,
      },
      auth: true,
    })
  } catch (error) {
    askError.value = error.message
  } finally {
    askLoading.value = false
  }
}

function candidateTitle(item) {
  return item?.title || item?.matched_question || `知识 #${item?.item_id || item?.id || ''}`
}

function candidateBody(item) {
  return item?.chunk_text || item?.answer || ''
}

async function loadKnowledge() {
  knowledgeLoading.value = true
  knowledgeError.value = ''

  try {
    const params = new URLSearchParams({
      page: String(knowledgePage.value),
      page_size: String(knowledgePageSize),
      status: 'approved',
    })
    const result = await requestJSON(`/api/v1/knowledge?${params}`)
    knowledgeItems.value = result.items || []
    knowledgeTotal.value = result.total || 0
  } catch (error) {
    knowledgeError.value = error.message
  } finally {
    knowledgeLoading.value = false
  }
}

function goToKnowledgePage(page) {
  if (page < 1 || page > totalKnowledgePages.value) return
  knowledgePage.value = page
  loadKnowledge()
}

async function loadUsers() {
  if (!isAdminLoggedIn.value) {
    userError.value = '请先登录管理员账号'
    return
  }

  userLoading.value = true
  userError.value = ''

  try {
    const params = new URLSearchParams({
      page: String(userPage.value),
      page_size: String(userPageSize),
    })
    const result = await requestJSON(`/api/v1/admin/users?${params}`, { auth: true })
    users.value = result.items || []
    userTotal.value = result.total || 0
  } catch (error) {
    if (!handleAdminAuthError(error)) {
      userError.value = error.message
    }
  } finally {
    userLoading.value = false
  }
}

function goToUserPage(page) {
  if (page < 1 || page > totalUserPages.value) return
  userPage.value = page
  loadUsers()
}

watch(activePanel, (panel) => {
  if (panel === 'review' && reviewEntered.value && isAdminLoggedIn.value) {
    loadSubmissions()
  }
  if (panel === 'knowledge') {
    loadKnowledge()
  }
  if (panel === 'users') {
    loadUsers()
  }
})

onMounted(() => {
  if (isAdminLoggedIn.value) {
    adminMessage.value = `已登录：${adminUser.value?.username || '管理员'}，请进入审核端`
  }
  loadKnowledge()
})
</script>

<template>
  <div class="console-layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1>Console</h1>
        <p>校园生活百事通</p>
      </div>
      <nav class="sidebar-nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          class="nav-item"
          :class="{ active: activePanel === item.key }"
          @click="activePanel = item.key"
        >
          {{ item.label }}
        </button>
      </nav>
      <div class="sidebar-footer">
        <t-tag theme="primary" variant="light" size="small">API {{ apiBaseURL }}</t-tag>
      </div>
    </aside>

    <main class="main-content">
      <!-- 审核面板 -->
      <section v-if="activePanel === 'review'" class="panel review-panel">
        <div class="panel-title">
          <h2>审核面板</h2>
          <span>{{ isAdminLoggedIn ? `管理员 ${adminUser?.username || ''}` : '请先登录管理员账号' }}</span>
        </div>

        <div v-if="!isAdminLoggedIn" class="login-row">
          <t-input
            v-model="adminForm.username"
            placeholder="管理员账号"
            :disabled="adminLoading"
          />
          <t-input
            v-model="adminForm.password"
            type="password"
            placeholder="管理员密码"
            :disabled="adminLoading"
            @keydown.enter.prevent="loginAdmin"
          />
          <t-button
            theme="primary"
            :loading="adminLoading"
            :disabled="!canLoginAdmin"
            @click="loginAdmin"
          >
            登录
          </t-button>
        </div>

        <div v-else-if="!reviewEntered" class="review-entry-actions">
          <p>当前管理员已登录，进入后可审核学生投稿并执行入库操作。</p>
          <div>
            <t-button
              theme="primary"
              :loading="enterReviewLoading"
              :disabled="!canEnterReview"
              @click="enterReviewPanel"
            >
              进入审核端
            </t-button>
            <t-button variant="text" :disabled="enterReviewLoading" @click="logoutAdmin">退出</t-button>
          </div>
        </div>

        <div v-else class="review-layout">
          <aside class="submission-list">
            <div class="submission-toolbar">
              <select
                v-model="reviewStatus"
                class="status-select"
                :disabled="submissionsLoading || isReviewBusy"
                @change="loadSubmissions"
              >
                <option
                  v-for="option in statusOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
              <t-button
                variant="outline"
                :loading="submissionsLoading"
                :disabled="submissionsLoading || isReviewBusy"
                @click="loadSubmissions"
              >
                刷新
              </t-button>
              <t-button variant="text" :disabled="isReviewBusy" @click="logoutAdmin">退出</t-button>
            </div>

            <div v-if="submissionsLoading" class="loading-state">
              正在加载投稿...
            </div>

            <div v-else-if="!submissions.length" class="empty-state">
              {{ emptySubmissionText }}
            </div>

            <button
              v-for="submission in submissions"
              :key="submission.id"
              class="submission-item"
              :class="{ active: selectedSubmission?.id === submission.id }"
              type="button"
              :disabled="isReviewBusy"
              @click="fillReviewForm(submission)"
            >
              <strong>{{ submission.question }}</strong>
              <span class="submission-summary">
                <t-tag
                  size="small"
                  variant="light"
                  :theme="statusTheme(submission.status)"
                >
                  {{ statusLabel(submission.status) }}
                </t-tag>
                #{{ submission.id }} · {{ formatDateTime(submission.created_at) }}
              </span>
            </button>
          </aside>

          <section class="review-detail">
            <div v-if="selectedSubmission" class="review-form">
              <div class="review-meta">
                <div>
                  <span class="review-meta-label">状态</span>
                  <t-tag variant="light" :theme="statusTheme(selectedSubmission.status)">
                    {{ statusLabel(selectedSubmission.status) }}
                  </t-tag>
                </div>
                <div>
                  <span class="review-meta-label">投稿编号</span>
                  <strong>#{{ selectedSubmission.id }}</strong>
                </div>
                <div>
                  <span class="review-meta-label">创建时间</span>
                  <strong>{{ formatDateTime(selectedSubmission.created_at) }}</strong>
                </div>
                <div>
                  <span class="review-meta-label">审核时间</span>
                  <strong>{{ formatDateTime(selectedSubmission.reviewed_at) }}</strong>
                </div>
                <div class="review-note-meta">
                  <span class="review-meta-label">审核备注</span>
                  <strong>{{ selectedSubmission.reviewer_note || '-' }}</strong>
                </div>
              </div>

              <t-form label-align="top" @submit.prevent>
                <t-form-item label="问题">
                  <t-textarea
                    v-model="reviewForm.question"
                    :autosize="{ minRows: 2, maxRows: 4 }"
                    :disabled="!canEditReviewForm"
                  />
                </t-form-item>
                <t-form-item label="答案">
                  <t-textarea
                    v-model="reviewForm.answer"
                    :autosize="{ minRows: 4, maxRows: 7 }"
                    :disabled="!canEditReviewForm"
                  />
                </t-form-item>

                <div class="form-row">
                  <t-form-item label="分类">
                    <t-input
                      v-model="reviewForm.category"
                      :disabled="!canEditReviewForm"
                    />
                  </t-form-item>
                  <t-form-item label="标签">
                    <t-input
                      v-model="reviewForm.tags"
                      placeholder="食堂，营业时间"
                      :disabled="!canEditReviewForm"
                    />
                  </t-form-item>
                </div>

                <div class="form-row">
                  <t-form-item label="来源">
                    <t-input
                      v-model="reviewForm.source"
                      :disabled="!canEditReviewForm"
                    />
                  </t-form-item>
                  <t-form-item label="备注">
                    <t-input
                      v-model="reviewForm.remark"
                      :disabled="!canEditReviewForm"
                    />
                  </t-form-item>
                </div>

                <t-form-item label="审核备注">
                  <t-input
                    v-model="reviewForm.reviewerNote"
                    placeholder="可填写通过或驳回原因"
                    :disabled="!canEditReviewForm"
                  />
                </t-form-item>

                <div class="review-actions">
                  <t-button
                    theme="success"
                    :loading="approveLoading"
                    :disabled="!canApproveSubmission"
                    @click="approveSubmission"
                  >
                    通过并入库
                  </t-button>
                  <t-button
                    theme="danger"
                    variant="outline"
                    :loading="rejectLoading"
                    :disabled="!canRejectSubmission"
                    @click="rejectSubmission"
                  >
                    驳回
                  </t-button>
                </div>
              </t-form>
            </div>

            <div v-else class="empty-state">请选择一条投稿</div>

            <t-alert
              v-if="adminMessage"
              class="feedback"
              theme="success"
              :message="adminMessage"
            />
            <t-alert
              v-if="adminError"
              class="feedback"
              theme="error"
              :message="adminError"
            />
            <t-alert
              v-if="reviewMessage"
              class="feedback"
              theme="success"
              :message="reviewMessage"
            />
            <t-alert
              v-if="reviewError"
              class="feedback"
              theme="error"
              :message="reviewError"
            />
          </section>
        </div>
      </section>

      <!-- 知识录入 -->
      <section v-if="activePanel === 'entry'" class="panel">
        <div class="panel-title">
          <h2>知识录入</h2>
          <span>保存时会生成向量并入库</span>
        </div>

        <t-form label-align="top" @submit.prevent>
          <t-form-item label="问题">
            <t-textarea
              v-model="knowledgeForm.question"
              placeholder="例如：三食堂晚上几点关门？"
              :autosize="{ minRows: 2, maxRows: 4 }"
              :disabled="submitLoading"
            />
          </t-form-item>

          <t-form-item label="答案">
            <t-textarea
              v-model="knowledgeForm.answer"
              placeholder="填写可以直接返回给用户的答案"
              :autosize="{ minRows: 5, maxRows: 8 }"
              :disabled="submitLoading"
            />
          </t-form-item>

          <div class="form-row">
            <t-form-item label="分类">
              <t-input
                v-model="knowledgeForm.category"
                placeholder="餐饮服务"
                :disabled="submitLoading"
              />
            </t-form-item>
            <t-form-item label="标签">
              <t-input
                v-model="knowledgeForm.tags"
                placeholder="食堂，营业时间，关门"
                :disabled="submitLoading"
              />
            </t-form-item>
          </div>

          <div class="form-row">
            <t-form-item label="来源">
              <t-input
                v-model="knowledgeForm.source"
                placeholder="后勤公告"
                :disabled="submitLoading"
              />
            </t-form-item>
            <t-form-item label="备注">
              <t-input
                v-model="knowledgeForm.remark"
                placeholder="可选"
                :disabled="submitLoading"
              />
            </t-form-item>
          </div>

          <t-button
            theme="primary"
            block
            :loading="submitLoading"
            :disabled="!canSubmitKnowledge"
            @click="submitKnowledge"
          >
            保存知识
          </t-button>
        </t-form>

        <t-alert
          v-if="submitMessage"
          class="feedback"
          theme="success"
          :message="submitMessage"
        />
        <t-alert
          v-if="submitError"
          class="feedback"
          theme="error"
          :message="submitError"
        />
      </section>

      <!-- 问答测试 -->
      <section v-if="activePanel === 'qa'" class="panel">
        <div class="panel-title">
          <h2>问答测试</h2>
          <span>请求返回前会锁定提问区</span>
        </div>

        <t-textarea
          v-model="askForm.question"
          placeholder="例如：食堂几点关门？"
          :autosize="{ minRows: 4, maxRows: 6 }"
          :disabled="askLoading"
          @keydown.enter.prevent="askQuestion"
        />

        <div class="ask-actions">
          <t-button
            theme="primary"
            :loading="askLoading"
            :disabled="!canAsk"
            @click="askQuestion"
          >
            提问
          </t-button>
        </div>

        <t-alert
          v-if="askError"
          class="feedback"
          theme="error"
          :message="askError"
        />

        <t-alert
          v-if="askResult && !askResult.answered"
          class="feedback"
          theme="warning"
          :message="`未找到足够相关的答案。最高相似度 ${Number(askResult.candidates?.[0]?.score || 0).toFixed(4)}，命中阈值 ${Number(askResult.min_score || 0).toFixed(2)}。`"
        />

        <t-alert
          v-if="askResult?.ai_error"
          class="feedback"
          theme="warning"
          :message="`AI 回答生成失败，已返回知识库原始结果：${askResult.ai_error}`"
        />

        <div v-if="askResult?.ai_answer" class="ai-answer-box">
          <div class="answer-meta">
            <t-tag theme="primary" variant="light">AI 回答</t-tag>
            <span>基于候选知识生成</span>
          </div>
          <p>{{ askResult.ai_answer }}</p>
        </div>

        <div v-if="askResult?.answer" class="answer-box">
          <div class="answer-meta">
            <t-tag theme="success" variant="light">
              {{ askResult.answer.category || '未分类' }}
            </t-tag>
            <span>相似度 {{ Number(askResult.answer.score || 0).toFixed(4) }}</span>
          </div>
          <h3>{{ candidateTitle(askResult.answer) }}</h3>
          <p>{{ candidateBody(askResult.answer) }}</p>
          <a
            v-if="askResult.answer.source_url"
            class="source-link"
            :href="askResult.answer.source_url"
            target="_blank"
            rel="noreferrer"
          >
            查看来源
          </a>
          <div v-if="askResult.answer.tags?.length" class="tag-list">
            <t-tag
              v-for="tag in askResult.answer.tags"
              :key="tag"
              variant="light"
            >
              {{ tag }}
            </t-tag>
          </div>
        </div>

        <div v-if="askResult?.candidates?.length" class="candidate-section">
          <div class="candidate-title">
            <h3>候选结果</h3>
            <span>相似度 = 1 - 余弦距离，低于 {{ Number(askResult.min_score || 0).toFixed(2) }} 不自动回答</span>
          </div>
          <div class="candidate-list">
            <div
              v-for="(item, index) in askResult.candidates"
              :key="item.chunk_id || item.id"
              class="candidate-item"
            >
              <div class="candidate-rank">{{ index + 1 }}</div>
              <div class="candidate-body">
                <div class="candidate-head">
                  <strong>{{ candidateTitle(item) }}</strong>
                  <span>{{ Number(item.score || 0).toFixed(4) }}</span>
                </div>
                <p>{{ candidateBody(item) }}</p>
                <div class="candidate-foot">
                  <t-tag size="small" variant="light">
                    {{ item.category || '未分类' }}
                  </t-tag>
                  <span v-if="item.chunk_id">片段 #{{ item.chunk_id }}</span>
                  <span v-if="item.tags?.length">{{ item.tags.join(' / ') }}</span>
                  <a
                    v-if="item.source_url"
                    class="source-link"
                    :href="item.source_url"
                    target="_blank"
                    rel="noreferrer"
                  >
                    来源
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 知识库浏览 -->
      <section v-if="activePanel === 'knowledge'" class="panel knowledge-browse-panel">
        <div class="panel-title">
          <div>
            <h2>知识库浏览</h2>
            <span>共 {{ knowledgeTotal }} 条已录入知识</span>
          </div>
          <t-button
            variant="outline"
            size="small"
            :loading="knowledgeLoading"
            @click="loadKnowledge"
          >
            刷新
          </t-button>
        </div>

        <div v-if="knowledgeLoading" class="loading-state">正在加载知识库...</div>

        <t-alert
          v-else-if="knowledgeError"
          class="feedback"
          theme="error"
          :message="knowledgeError"
        />

        <div v-else-if="!knowledgeItems.length" class="empty-state">
          知识库中暂无已录入的知识
        </div>

        <template v-else>
          <div class="knowledge-table-wrap">
            <table class="knowledge-table">
              <thead>
                <tr>
                  <th class="col-id">#</th>
                  <th class="col-question">问题</th>
                  <th class="col-answer">答案</th>
                  <th class="col-category">分类</th>
                  <th class="col-tags">标签</th>
                  <th class="col-date">录入时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in knowledgeItems" :key="item.id">
                  <td class="col-id">{{ item.id }}</td>
                  <td class="col-question">{{ item.question }}</td>
                  <td class="col-answer">{{ item.answer }}</td>
                  <td class="col-category">
                    <t-tag v-if="item.category" size="small" variant="light">{{ item.category }}</t-tag>
                    <span v-else>-</span>
                  </td>
                  <td class="col-tags">
                    <t-tag v-for="tag in item.tags" :key="tag" size="small" variant="outline" class="tag-chip">{{ tag }}</t-tag>
                    <span v-if="!item.tags?.length">-</span>
                  </td>
                  <td class="col-date">{{ formatDateTime(item.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="knowledge-pagination">
            <span class="pagination-info">第 {{ knowledgePage }} / {{ totalKnowledgePages }} 页，共 {{ knowledgeTotal }} 条</span>
            <div class="pagination-jump">
              <span>跳至</span>
              <input
                type="number"
                :min="1"
                :max="totalKnowledgePages"
                v-model.number="knowledgeJumpPage"
                @keydown.enter="goToKnowledgePage(knowledgeJumpPage)"
                class="jump-input"
              />
              <span>页</span>
              <t-button
                size="small"
                variant="outline"
                :disabled="!knowledgeJumpPage || knowledgeJumpPage < 1 || knowledgeJumpPage > totalKnowledgePages"
                @click="goToKnowledgePage(knowledgeJumpPage)"
              >
                GO
              </t-button>
            </div>
            <div class="pagination-btns">
              <t-button
                variant="outline"
                size="small"
                :disabled="knowledgePage <= 1"
                @click="goToKnowledgePage(knowledgePage - 1)"
              >
                上一页
              </t-button>
              <t-button
                variant="outline"
                size="small"
                :disabled="knowledgePage >= totalKnowledgePages"
                @click="goToKnowledgePage(knowledgePage + 1)"
              >
                下一页
              </t-button>
            </div>
          </div>
        </template>
      </section>

      <!-- 用户管理 -->
      <section v-if="activePanel === 'users'" class="panel knowledge-browse-panel">
        <div class="panel-title">
          <div>
            <h2>用户管理</h2>
            <span>共 {{ userTotal }} 位注册用户</span>
          </div>
          <t-button
            variant="outline"
            size="small"
            :loading="userLoading"
            @click="loadUsers"
          >
            刷新
          </t-button>
        </div>

        <div v-if="userLoading" class="loading-state">正在加载用户列表...</div>

        <t-alert
          v-else-if="userError"
          class="feedback"
          theme="error"
          :message="userError"
        />

        <div v-else-if="!users.length" class="empty-state">
          暂无注册用户
        </div>

        <template v-else>
          <div class="knowledge-table-wrap">
            <table class="knowledge-table">
              <thead>
                <tr>
                  <th class="col-id">ID</th>
                  <th class="col-username">用户名</th>
                  <th class="col-nickname">昵称</th>
                  <th class="col-date">注册时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users" :key="user.id">
                  <td class="col-id">{{ user.id }}</td>
                  <td class="col-username">{{ user.username }}</td>
                  <td class="col-nickname">{{ user.nickname || '-' }}</td>
                  <td class="col-date">{{ formatDateTime(user.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="knowledge-pagination">
            <span class="pagination-info">第 {{ userPage }} / {{ totalUserPages }} 页，共 {{ userTotal }} 条</span>
            <div class="pagination-jump">
              <span>跳至</span>
              <input
                type="number"
                :min="1"
                :max="totalUserPages"
                v-model.number="userJumpPage"
                @keydown.enter="goToUserPage(userJumpPage)"
                class="jump-input"
              />
              <span>页</span>
              <t-button
                size="small"
                variant="outline"
                :disabled="!userJumpPage || userJumpPage < 1 || userJumpPage > totalUserPages"
                @click="goToUserPage(userJumpPage)"
              >
                GO
              </t-button>
            </div>
            <div class="pagination-btns">
              <t-button
                variant="outline"
                size="small"
                :disabled="userPage <= 1"
                @click="goToUserPage(userPage - 1)"
              >
                上一页
              </t-button>
              <t-button
                variant="outline"
                size="small"
                :disabled="userPage >= totalUserPages"
                @click="goToUserPage(userPage + 1)"
              >
                下一页
              </t-button>
            </div>
          </div>
        </template>
      </section>
    </main>
  </div>
</template>
