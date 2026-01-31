<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Test Cases</h2>
      <div>
        <button class="btn btn-danger me-2" v-if="selectedIds.length > 0" @click="batchDelete">Delete Selected ({{ selectedIds.length }})</button>
        <button class="btn btn-primary" @click="openDialog()">Add Test Case</button>
        <button class="btn btn-secondary ms-2" @click="openGenerateDialog()">Generate with AI</button>
      </div>
    </div>

    <div class="table-responsive">
      <table class="table table-striped table-hover">
        <thead class="table-dark">
          <tr>
            <th scope="col" style="width: 40px;">
              <input type="checkbox" class="form-check-input" :checked="isAllSelected" @change="toggleSelectAll">
            </th>
            <th scope="col" style="width: 1%; white-space: nowrap;">ID</th>
            <th scope="col">Input</th>
            <th scope="col" style="width: 1%; white-space: nowrap;">MD5</th>
            <th scope="col" style="width: 1%; white-space: nowrap;">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="testCase in testCases" :key="testCase.id">
            <td>
              <input type="checkbox" class="form-check-input" :value="testCase.id" v-model="selectedIds">
            </td>
            <td class="text-nowrap">{{ testCase.id }}</td>
            <td>
              <div class="text-wrap" :title="testCase.input" style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden;">{{ testCase.input }}</div>
            </td>
            <td class="text-nowrap">
              <small class="text-muted">{{ testCase.input_md5 }}</small>
            </td>
            <td class="text-nowrap">
              <button class="btn btn-sm btn-outline-primary me-2" @click="openDialog(testCase)">Edit</button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteTestCase(testCase.id)">Delete</button>
            </td>
          </tr>
          <tr v-if="testCases.length === 0">
            <td colspan="5" class="text-center">No test cases found.</td>
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

    <!-- Modal -->
    <div class="modal fade" id="testCaseModal" tabindex="-1" aria-hidden="true" ref="modalRef">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? 'Edit Test Case' : 'Add Test Case' }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveTestCase">
              <div class="mb-3">
                <label class="form-label">Input Data</label>
                <textarea class="form-control" v-model="form.input" rows="4" required></textarea>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="saveTestCase">Confirm</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Generate Modal -->
    <div class="modal fade" id="generateModal" tabindex="-1" aria-hidden="true" ref="generateModalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Generate Test Cases with AI</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="generateTestCases">
              <div class="mb-3">
                <label class="form-label">LLM Config</label>
                <select class="form-select" v-model="generateForm.config_id" required>
                  <option v-for="config in llmConfigs" :key="config.id" :value="config.id">{{ config.name }}</option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Prompts (Select Multiple)</label>
                <div class="border rounded p-2" style="max-height: 200px; overflow-y: auto;">
                  <div v-for="prompt in prompts" :key="prompt.id" class="form-check">
                    <input class="form-check-input" type="checkbox" :value="prompt.id" v-model="generateForm.prompt_ids" :id="'prompt-' + prompt.id">
                    <label class="form-check-label" :for="'prompt-' + prompt.id">
                      {{ prompt.name }}
                    </label>
                  </div>
                </div>
                <div v-if="generateForm.prompt_ids.length === 0" class="text-danger small mt-1">Please select at least one prompt</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Count (Per Prompt)</label>
                <input type="number" class="form-control" v-model="generateForm.count" min="1" max="10" required>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="generateTestCases" :disabled="isGenerating">
              <span v-if="isGenerating" class="spinner-border spinner-border-sm me-1"></span>
              Generate
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed, inject } from 'vue'
import axios from 'axios'
import { Modal } from 'bootstrap'
import { useProject } from '../composables/useProject'

const showModal = inject('showGlobalModal')
const showConfirm = inject('showGlobalConfirm')

const testCases = ref([])
const selectedIds = ref([])
const currentPage = ref(1)
const pageSize = ref(30)
const totalItems = ref(0)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))

const form = ref({
  id: null,
  name: '',
  input: '',
  expected_output: '',
  tags: ''
})
const modalRef = ref(null)
let modalInstance = null

const llmConfigs = ref([])
const prompts = ref([])
const generateForm = ref({
  config_id: '',
  prompt_ids: [],
  count: 5
})
const generateModalRef = ref(null)
let generateModalInstance = null
const isGenerating = ref(false)
const { currentProjectId } = useProject()

const isAllSelected = computed(() => {
  return testCases.value.length > 0 && selectedIds.value.length === testCases.value.length
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = testCases.value.map(tc => tc.id)
  }
}

const exportJson = () => {
  const dataToExport = selectedIds.value.length > 0 
    ? testCases.value.filter(tc => selectedIds.value.includes(tc.id))
    : testCases.value
    
  if (dataToExport.length === 0) {
    showModal('Info', 'No test cases to export')
    return
  }

  // Export minimal format as requested (only input?)
  // But JSON export might want full data. User said "modules" are not needed.
  // Let's keep existing export but it will have empty fields for removed cols.
  const dataStr = JSON.stringify(dataToExport, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `test_cases_${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const batchDelete = async () => {
  if (!await showConfirm('Confirm', `Are you sure you want to delete ${selectedIds.value.length} test cases?`)) return
  try {
    await axios.delete('/api/test-cases/batch', { data: { ids: selectedIds.value } })
    selectedIds.value = []
    fetchTestCases()
  } catch (error) {
    showModal('Error', 'Failed to delete selected test cases')
  }
}

const fetchTestCases = async () => {
  if (!currentProjectId.value) {
    testCases.value = []
    totalItems.value = 0
    selectedIds.value = []
    return
  }
  try {
    const res = await axios.get('/api/test-cases', {
      params: { 
        project_id: currentProjectId.value,
        page: currentPage.value,
        page_size: pageSize.value
      }
    })
    if (res.data && res.data.items !== undefined) {
      testCases.value = res.data.items || []
      totalItems.value = res.data.total || 0
    } else if (Array.isArray(res.data)) {
      testCases.value = res.data
      totalItems.value = res.data.length
    } else {
      testCases.value = []
      totalItems.value = 0
    }
    selectedIds.value = [] // Reset selection on fetch
  } catch (error) {
    showModal('Error', 'Failed to fetch test cases')
  }
}

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchTestCases()
}

const openDialog = (row) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = { id: null, name: '', input: '', expected_output: '', tags: '' }
  }
  modalInstance.show()
}

const saveTestCase = async () => {
  try {
    // Fill required fields that are hidden
    if (!form.value.name) form.value.name = 'Manual Test Case'
    if (!form.value.project_id) form.value.project_id = currentProjectId.value ? parseInt(currentProjectId.value) : 0
    
    if (form.value.id) {
      await axios.put(`/api/test-cases/${form.value.id}`, form.value)
    } else {
      await axios.post('/api/test-cases', form.value)
    }
    modalInstance.hide()
    fetchTestCases()
  } catch (error) {
    showModal('Error', 'Failed to save: ' + (error.response?.data?.error || error.message))
  }
}

const deleteTestCase = async (id) => {
  if (!await showConfirm('Confirm', 'Are you sure you want to delete this test case?')) return
  try {
    await axios.delete(`/api/test-cases/${id}`)
    fetchTestCases()
  } catch (error) {
    showModal('Error', 'Failed to delete')
  }
}

const fetchConfigs = async () => {
  try {
    const res = await axios.get('/api/llm-configs', { params: { page_size: 100 } })
    llmConfigs.value = res.data.items || res.data
  } catch (error) {
    console.error('Failed to fetch configs')
  }
}

const fetchPrompts = async () => {
  if (!currentProjectId.value) {
    prompts.value = []
    return
  }
  try {
    const res = await axios.get('/api/prompts', { 
      params: { project_id: currentProjectId.value, page_size: 100 } 
    })
    prompts.value = res.data.items || res.data
  } catch (error) {
    console.error('Failed to fetch prompts')
  }
}

watch(currentProjectId, (newId) => {
  if (newId) {
    currentPage.value = 1
    fetchTestCases()
    fetchPrompts()
  } else {
    testCases.value = []
    totalItems.value = 0
    prompts.value = []
  }
})

const openGenerateDialog = () => {
  generateForm.value = { config_id: '', prompt_ids: [], count: 5 }
  const defaultConfig = llmConfigs.value.find(c => c.is_default)
  if (defaultConfig) {
    generateForm.value.config_id = defaultConfig.id
  } else if (llmConfigs.value.length > 0) {
    generateForm.value.config_id = llmConfigs.value[0].id
  }
  
  // Auto-select first prompt if available? Maybe not for multi-select.
  // User can select what they want.
  generateModalInstance.show()
}

const generateTestCases = async () => {
  if (!generateForm.value.config_id || generateForm.value.prompt_ids.length === 0) return
  isGenerating.value = true
  try {
    await axios.post('/api/test-cases/generate', {
      config_id: generateForm.value.config_id,
      prompt_ids: generateForm.value.prompt_ids,
      count: generateForm.value.count
    })
    generateModalInstance.hide()
    fetchTestCases()
    showModal('Success', 'Test cases generated successfully!')
  } catch (error) {
    showModal('Error', 'Failed to generate test cases: ' + (error.response?.data?.error || error.message))
  } finally {
    isGenerating.value = false
  }
}

onMounted(() => {
  fetchTestCases()
  fetchConfigs()
  fetchPrompts()
  modalInstance = new Modal(modalRef.value)
  generateModalInstance = new Modal(generateModalRef.value)
})
</script>
