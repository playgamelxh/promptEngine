<template>
  <div>
    <h2 class="mb-4">LLM Test Case Management</h2>

    <div class="row mb-4">
      <div class="col-md-5">
        <label class="form-label">Select Prompt</label>
        <select class="form-select" v-model="selectedPromptId" @change="fetchResults">
          <option value="" disabled>Choose a prompt...</option>
          <option v-for="prompt in prompts" :key="prompt.id" :value="prompt.id">
            {{ prompt.name }}
          </option>
        </select>
      </div>
      <div class="col-md-5">
        <label class="form-label">Select LLM Config</label>
        <select class="form-select" v-model="selectedConfigId">
          <option value="" disabled>Choose a config...</option>
          <option v-for="config in llmConfigs" :key="config.id" :value="config.id">
            {{ config.name }}
          </option>
        </select>
      </div>
      <div class="col-md-2 d-flex align-items-end">
        <button class="btn btn-primary w-100" @click="runEvaluation" :disabled="!selectedPromptId || !selectedConfigId || isRunning">
          <span v-if="isRunning" class="spinner-border spinner-border-sm me-1"></span>
          {{ isRunning ? 'Running...' : 'Run Evaluation' }}
        </button>
      </div>
    </div>

    <!-- Task Status -->
    <div class="row mb-4" v-if="isRunning || taskStatus.status === 'running'">
      <div class="col-12">
        <div class="alert alert-info d-flex justify-content-between align-items-center">
           <div class="flex-grow-1 me-3">
             <div class="d-flex justify-content-between mb-1">
               <strong>Running Task:</strong>
               <span>{{ taskStatus.progress }} / {{ taskStatus.total }}</span>
             </div>
             <div class="progress" style="height: 20px;">
                <div class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" 
                     :style="{ width: (taskStatus.total > 0 ? (taskStatus.progress / taskStatus.total * 100) : 0) + '%' }">
                     {{ taskStatus.total > 0 ? Math.round(taskStatus.progress / taskStatus.total * 100) : 0 }}%
                </div>
             </div>
             <div class="text-muted small text-truncate mt-1">{{ taskStatus.message }}</div>
           </div>
           <button class="btn btn-sm btn-danger" @click="stopEvaluation">Stop</button>
         </div>
      </div>
    </div>

    <!-- Statistics -->
    <div class="row mb-4" v-if="results.length > 0">
      <div class="col-md-3">
        <div class="card text-center bg-light">
          <div class="card-body">
            <h5 class="card-title">Total</h5>
            <p class="card-text display-6">{{ results.length }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card text-center text-white bg-success">
          <div class="card-body">
            <h5 class="card-title">Passed</h5>
            <p class="card-text display-6">{{ passCount }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card text-center text-white bg-danger">
          <div class="card-body">
            <h5 class="card-title">Failed</h5>
            <p class="card-text display-6">{{ failCount }}</p>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card text-center bg-info text-white">
          <div class="card-body">
            <h5 class="card-title">Pass Rate</h5>
            <p class="card-text display-6">{{ passRate }}%</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Results Table -->
    <div class="table-responsive">
      <table class="table table-striped table-hover align-middle">
        <thead class="table-dark">
          <tr>
            <th scope="col" style="width: 50px;">
              <input type="checkbox" class="form-check-input" :checked="isAllSelected" @change="toggleSelectAll">
            </th>
            <th scope="col">Input</th>
            <th scope="col">Output</th>
            <th scope="col">Evaluation</th>
            <th scope="col" style="width: 120px;">Status</th>
            <th scope="col" style="width: 100px;">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="result in results" :key="result.id">
            <td>
              <input type="checkbox" class="form-check-input" :value="result.id" v-model="selectedIds">
            </td>
            <td>
              <div class="text-wrap" style="max-width: 300px; max-height: 100px; overflow-y: auto;">{{ result.input }}</div>
            </td>
            <td>
              <div class="text-wrap" style="max-width: 300px; max-height: 100px; overflow-y: auto;">{{ result.output }}</div>
            </td>
            <td>
              <div class="text-wrap small" style="max-width: 300px; max-height: 100px; overflow-y: auto;">{{ result.evaluation }}</div>
            </td>
            <td>
              <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" :id="'status-'+result.id" :checked="result.is_pass" @change="toggleStatus(result)">
                <label class="form-check-label" :for="'status-'+result.id">
                  {{ result.is_pass ? 'Pass' : 'Fail' }}
                </label>
              </div>
            </td>
            <td>
              <button class="btn btn-sm btn-outline-primary me-1" @click="reEvaluate([result.id])" :disabled="!selectedConfigId || isRunning">Eval</button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteResult(result.id)" :disabled="isRunning">Delete</button>
            </td>
          </tr>
          <tr v-if="results.length === 0">
            <td colspan="6" class="text-center">No results found. Select a prompt to view or run tests.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="d-flex justify-content-between align-items-center mt-3" v-if="totalItems > 0">
      <div>
        Showing {{ (currentPage - 1) * pageSize + 1 }} to {{ Math.min(currentPage * pageSize, totalItems) }} of {{ totalItems }} entries
      </div>
      <nav aria-label="Page navigation">
        <ul class="pagination mb-0">
          <li class="page-item" :class="{ disabled: currentPage === 1 }">
            <button class="page-link" @click="changePage(currentPage - 1)">Previous</button>
          </li>
          <li class="page-item" :class="{ disabled: currentPage >= totalPages }">
            <button class="page-link" @click="changePage(currentPage + 1)">Next</button>
          </li>
        </ul>
      </nav>
    </div>

    <div class="mt-3" v-if="selectedIds.length > 0">
      <button class="btn btn-primary me-2" @click="reEvaluate(selectedIds)" :disabled="!selectedConfigId || isRunning">Re-evaluate Selected ({{ selectedIds.length }})</button>
      <button class="btn btn-danger" @click="batchDelete" :disabled="isRunning">Delete Selected ({{ selectedIds.length }})</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'
import axios from 'axios'
import { useProject } from '../composables/useProject'
import { useTask } from '../composables/useTask'

const showModal = inject('showGlobalModal')
const showConfirm = inject('showGlobalConfirm')

const { currentProjectId } = useProject()
const { currentTaskId, isRunning, taskStatus, startTask, stopRunningTask } = useTask()

const prompts = ref([])
const llmConfigs = ref([])
const results = ref([])
const selectedPromptId = ref('')
const selectedConfigId = ref('')
// Local task state removed in favor of useTask
const selectedIds = ref([])
const currentPage = ref(1)
const pageSize = ref(30)
const totalItems = ref(0)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))

const passCount = computed(() => results.value.filter(r => r.is_pass).length)
const failCount = computed(() => results.value.length - passCount.value)
const passRate = computed(() => {
  if (results.value.length === 0) return 0
  return Math.round((passCount.value / results.value.length) * 100)
})

const isAllSelected = computed(() => {
  return results.value.length > 0 && selectedIds.value.length === results.value.length
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = results.value.map(r => r.id)
  }
}

const fetchData = async () => {
  if (!currentProjectId.value) return
  try {
    const [promptsRes, configsRes] = await Promise.all([
      axios.get(`/api/prompts`, { params: { project_id: currentProjectId.value, page_size: 100 } }),
      axios.get('/api/llm-configs', { params: { page_size: 100 } })
    ])
    prompts.value = promptsRes.data.items || promptsRes.data
    llmConfigs.value = configsRes.data.items || configsRes.data
    
    // Auto-select first if available
    if (prompts.value.length > 0 && !selectedPromptId.value) {
      selectedPromptId.value = prompts.value[0].id
      fetchResults()
    }
    if (llmConfigs.value.length > 0 && !selectedConfigId.value) {
      const defaultConfig = llmConfigs.value.find(c => c.is_default)
      if (defaultConfig) {
        selectedConfigId.value = defaultConfig.id
      } else {
        selectedConfigId.value = llmConfigs.value[0].id
      }
    }
  } catch (error) {
    console.error('Failed to fetch initial data', error)
  }
}

const fetchResults = async () => {
  if (!selectedPromptId.value) return
  try {
    const res = await axios.get('/api/llm-test-cases', {
      params: { 
        prompt_id: selectedPromptId.value,
        page: currentPage.value,
        page_size: pageSize.value
      }
    })
    if (res.data.items) {
      results.value = res.data.items
      totalItems.value = res.data.total
    } else {
      results.value = res.data
      totalItems.value = res.data.length
    }
    selectedIds.value = []
  } catch (error) {
    console.error('Failed to fetch results', error)
    results.value = []
  }
}

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchResults()
}

const reEvaluate = async (ids) => {
  if (!selectedConfigId.value || ids.length === 0) return
  
  try {
    const res = await axios.post('/api/llm-test-cases/evaluate', {
      test_case_ids: ids,
      config_id: selectedConfigId.value
    })
    
    if (res.data.task_id) {
      startTask(res.data.task_id)
      selectedIds.value = []
    }
  } catch (error) {
     if (error.response?.status === 409) {
        showModal('Error', 'Task is already running.')
     } else {
        showModal('Error', 'Failed to start evaluation: ' + (error.response?.data?.error || error.message))
     }
  }
}

const runEvaluation = async () => {
  if (!selectedPromptId.value || !selectedConfigId.value) return
  
  try {
    const res = await axios.post('/api/llm-test-cases/run-from-definitions', {
      prompt_id: selectedPromptId.value,
      config_id: selectedConfigId.value
    })
    
    if (res.data.task_id) {
      startTask(res.data.task_id)
      selectedIds.value = []
    }
  } catch (error) {
     if (error.response?.status === 409) {
        // Task already running, try to recover session if possible or just alert
        showModal('Error', 'Task is already running.')
     } else {
        showModal('Error', 'Failed to start evaluation: ' + (error.response?.data?.error || error.message))
     }
  }
}

// Watch for task status changes to update results and show notifications
watch(() => taskStatus.value.status, (newStatus, oldStatus) => {
  // Only react if we are transitioning from running/pending to a terminal state
  const wasRunning = oldStatus === 'running' || oldStatus === 'pending' || oldStatus === ''
  const isTerminal = ['completed', 'failed', 'stopped'].includes(newStatus)
  
  if (wasRunning && isTerminal) {
    fetchResults()
    if (newStatus === 'completed') {
      showModal('Task Completed', 'Evaluation finished!')
    } else if (newStatus === 'stopped') {
      showModal('Task Stopped', 'Evaluation stopped.')
    } else if (newStatus === 'failed') {
      showModal('Task Failed', 'Evaluation failed: ' + (taskStatus.value.error || taskStatus.value.message))
    }
  }
})

const stopEvaluation = async () => {
  if (!await showConfirm('Confirm', 'Are you sure you want to stop the evaluation?')) return
  if (!currentTaskId.value) return
  
  try {
    await stopRunningTask()
  } catch (error) {
    showModal('Error', 'Failed to stop task: ' + (error.response?.data?.error || error.message))
  }
}

const toggleStatus = async (result) => {
  const newStatus = !result.is_pass
  try {
    await axios.put(`/api/llm-test-cases/${result.id}`, {
      is_pass: newStatus,
      evaluation: result.evaluation // Keep existing evaluation text
    })
    result.is_pass = newStatus
  } catch (error) {
    showModal('Error', 'Failed to update status')
    // Revert visual change if failed (though checkbox toggles automatically, might need force update)
    result.is_pass = !newStatus 
  }
}

const deleteResult = async (id) => {
  // Use global confirm instead of native confirm
  if (!await showConfirm('Confirm', 'Delete this result?')) return
  try {
    await axios.delete(`/api/llm-test-cases/${id}`)
    results.value = results.value.filter(r => r.id !== id)
  } catch (error) {
    showModal('Error', 'Failed to delete')
  }
}

const batchDelete = async () => {
  if (!await showConfirm('Confirm', `Delete ${selectedIds.value.length} results?`)) return
  try {
    await axios.delete('/api/llm-test-cases/batch', { data: { ids: selectedIds.value } })
    results.value = results.value.filter(r => !selectedIds.value.includes(r.id))
    selectedIds.value = []
  } catch (error) {
    showModal('Error', 'Failed to batch delete')
  }
}

watch(currentProjectId, () => {
  selectedPromptId.value = ''
  results.value = []
  fetchData()
})

onMounted(() => {
  fetchData()
  // Task polling is handled globally by useTask/App.vue
})
</script>