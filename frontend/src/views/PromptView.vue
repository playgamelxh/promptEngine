<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Prompts</h2>
      <div>
        <button class="btn btn-info me-2" v-if="selectedIds.length > 0" @click="openBatchTestDialog">Generate Tests ({{ selectedIds.length }})</button>
        <button class="btn btn-danger me-2" v-if="selectedIds.length > 0" @click="batchDelete">Delete Selected ({{ selectedIds.length }})</button>
        <button class="btn btn-secondary me-2" @click="exportPrompts">Export JSON</button>
        <button class="btn btn-secondary me-2" @click="openGenerateDialog">Generate with AI</button>
        <button class="btn btn-primary" @click="openDialog()">Add Prompt</button>
      </div>
    </div>

    <div class="table-responsive">
      <table class="table table-striped table-hover">
        <thead class="table-dark">
          <tr>
            <th scope="col" style="width: 40px;">
              <input type="checkbox" class="form-check-input" :checked="isAllSelected" @change="toggleSelectAll">
            </th>
            <th scope="col">ID</th>
            <th scope="col">Name</th>
            <th scope="col" style="width: 40%;">Content</th>
            <th scope="col">Tags</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="prompt in prompts" :key="prompt.id">
            <td>
              <input type="checkbox" class="form-check-input" :value="prompt.id" v-model="selectedIds">
            </td>
            <td>{{ prompt.id }}</td>
            <td>{{ prompt.name }}</td>
            <td>
              <div class="text-wrap" :title="prompt.content" style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; max-height: 4.5em;">{{ prompt.content }}</div>
            </td>
            <td>
              <span v-for="tag in prompt.tags ? prompt.tags.split(',') : []" :key="tag" class="badge bg-secondary me-1">{{ tag.trim() }}</span>
            </td>
            <td>
              <button class="btn btn-sm btn-outline-primary me-2" @click="openDialog(prompt)">Edit</button>
              <button class="btn btn-sm btn-outline-danger" @click="deletePrompt(prompt.id)">Delete</button>
            </td>
          </tr>
          <tr v-if="prompts.length === 0">
            <td colspan="6" class="text-center">No prompts found.</td>
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
    <div class="modal fade" id="promptModal" tabindex="-1" aria-hidden="true" ref="modalRef">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? 'Edit Prompt' : 'Add Prompt' }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="savePrompt">
              <div class="mb-3">
                <label class="form-label">Name</label>
                <input type="text" class="form-control" v-model="form.name" required>
              </div>
              <div class="mb-3">
                <label class="form-label">Content</label>
                <textarea class="form-control" v-model="form.content" rows="6" required></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">Tags</label>
                <input type="text" class="form-control" v-model="form.tags" placeholder="Comma separated">
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="savePrompt">Confirm</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Generate Prompts Modal -->
    <div class="modal fade" id="generateModal" tabindex="-1" aria-hidden="true" ref="generateModalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Generate Prompts with AI</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="generatePrompts">
              <div class="mb-3">
                <label class="form-label">LLM Config</label>
                <select class="form-select" v-model="generateForm.config_id" required>
                  <option value="" disabled>Select Config</option>
                  <option v-for="config in llmConfigs" :key="config.id" :value="config.id">
                    {{ config.name }} ({{ config.model_name }})
                  </option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Instruction / Topic</label>
                <textarea class="form-control" v-model="generateForm.instruction" rows="3" placeholder="E.g., Create 5 prompts for a coding assistant specialized in Go." required></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">Count</label>
                <input type="number" class="form-control" v-model="generateForm.count" min="1" max="10" required>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="generatePrompts" :disabled="isGenerating">
              <span v-if="isGenerating" class="spinner-border spinner-border-sm me-1"></span>
              {{ isGenerating ? 'Generating...' : 'Generate' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Batch Test Generation Modal -->
    <div class="modal fade" id="batchTestModal" tabindex="-1" aria-hidden="true" ref="batchTestModalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Generate Test Cases for Selected Prompts</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="batchGenerateTests">
              <div class="mb-3">
                <label class="form-label">LLM Config</label>
                <select class="form-select" v-model="batchTestForm.config_id" required>
                  <option value="" disabled>Select Config</option>
                  <option v-for="config in llmConfigs" :key="config.id" :value="config.id">
                    {{ config.name }} ({{ config.model_name }})
                  </option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Count (per prompt)</label>
                <input type="number" class="form-control" v-model="batchTestForm.count" min="1" max="5" required>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="batchGenerateTests" :disabled="isBatchGenerating">
              <span v-if="isBatchGenerating" class="spinner-border spinner-border-sm me-1"></span>
              {{ isBatchGenerating ? 'Generating...' : 'Generate' }}
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

const prompts = ref([])
const llmConfigs = ref([])
const selectedIds = ref([])
const currentPage = ref(1)
const pageSize = ref(30)
const totalItems = ref(0)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))
const form = ref({
  id: null,
  name: '',
  content: '',
  tags: '',
  project_id: null
})
const generateForm = ref({
  config_id: '',
  instruction: '',
  count: 3
})
const batchTestForm = ref({
  config_id: '',
  count: 3
})

const modalRef = ref(null)
const generateModalRef = ref(null)
const batchTestModalRef = ref(null)
let modalInstance = null
let generateModalInstance = null
let batchTestModalInstance = null
const isGenerating = ref(false)
const isBatchGenerating = ref(false)

const { currentProjectId } = useProject()

const isAllSelected = computed(() => {
  return prompts.value.length > 0 && selectedIds.value.length === prompts.value.length
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = prompts.value.map(p => p.id)
  }
}

const batchDelete = async () => {
  if (!await showConfirm('Confirm', `Are you sure you want to delete ${selectedIds.value.length} prompts?`)) return
  try {
    await axios.delete('/api/prompts/batch', { data: { ids: selectedIds.value } })
    selectedIds.value = []
    fetchPrompts()
  } catch (error) {
    showModal('Error', 'Failed to delete selected prompts')
  }
}

const openBatchTestDialog = () => {
  batchTestForm.value = { config_id: '', count: 3 }
  const defaultConfig = llmConfigs.value.find(c => c.is_default)
  if (defaultConfig) {
    batchTestForm.value.config_id = defaultConfig.id
  }
  batchTestModalInstance.show()
}

const batchGenerateTests = async () => {
  if (!batchTestForm.value.config_id) {
    showModal('Warning', 'Please select LLM Config')
    return
  }
  isBatchGenerating.value = true
  let successCount = 0
  let failCount = 0

  for (const promptId of selectedIds.value) {
    try {
      await axios.post('/api/test-cases/generate', {
        prompt_id: promptId,
        config_id: batchTestForm.value.config_id,
        count: batchTestForm.value.count
      })
      successCount++
    } catch (error) {
      failCount++
      console.error(`Failed to generate tests for prompt ${promptId}:`, error)
    }
  }
  isBatchGenerating.value = false
  batchTestModalInstance.hide()
  showModal('Info', `Generation complete. Success: ${successCount}, Failed: ${failCount}. Check Test Cases page.`)
  selectedIds.value = []
}

const fetchPrompts = async () => {
  if (!currentProjectId.value) {
    prompts.value = []
    totalItems.value = 0
    return
  }
  try {
    const res = await axios.get('/api/prompts', {
      params: { 
        project_id: currentProjectId.value,
        page: currentPage.value,
        page_size: pageSize.value
      }
    })
    if (res.data && res.data.items !== undefined) {
      prompts.value = res.data.items || []
      totalItems.value = res.data.total || 0
    } else if (Array.isArray(res.data)) {
      prompts.value = res.data
      totalItems.value = res.data.length
    } else {
      prompts.value = []
      totalItems.value = 0
    }
    selectedIds.value = []
  } catch (error) {
    showModal('Error', 'Failed to fetch prompts')
  }
}

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchPrompts()
}

const fetchLLMConfigs = async () => {
  try {
    const res = await axios.get('/api/llm-configs')
    llmConfigs.value = res.data.items || res.data
  } catch (error) {
    console.error('Failed to fetch LLM configs')
  }
}

watch(currentProjectId, (newId) => {
  if (newId) {
    currentPage.value = 1
    fetchPrompts()
  } else {
    prompts.value = []
    totalItems.value = 0
  }
})

const openDialog = (row) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = { id: null, name: '', content: '', tags: '', project_id: parseInt(currentProjectId.value) }
  }
  modalInstance.show()
}

const openGenerateDialog = () => {
  generateForm.value = { config_id: '', instruction: '', count: 3 }
  const defaultConfig = llmConfigs.value.find(c => c.is_default)
  if (defaultConfig) {
    generateForm.value.config_id = defaultConfig.id
  }
  generateModalInstance.show()
}

const savePrompt = async () => {
  try {
    form.value.project_id = parseInt(currentProjectId.value)
    if (form.value.id) {
      await axios.put(`/api/prompts/${form.value.id}`, form.value)
    } else {
      await axios.post('/api/prompts', form.value)
    }
    modalInstance.hide()
    fetchPrompts()
  } catch (error) {
    showModal('Error', 'Failed to save')
  }
}

const deletePrompt = async (id) => {
  if (!await showConfirm('Confirm', 'Are you sure you want to delete this prompt?')) return
  try {
    await axios.delete(`/api/prompts/${id}`)
    fetchPrompts()
  } catch (error) {
    showModal('Error', 'Failed to delete')
  }
}

const exportPrompts = () => {
  const dataStr = JSON.stringify(prompts.value, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `prompts_project_${currentProjectId.value}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const generatePrompts = async () => {
  if (!generateForm.value.config_id || !generateForm.value.instruction) {
      showModal('Warning', 'Please fill in all fields')
      return
    }
    if (!currentProjectId.value) {
      showModal('Warning', 'Please select a project first')
      return
    }

  isGenerating.value = true
  try {
    await axios.post('/api/prompts/generate', {
      config_id: generateForm.value.config_id,
      instruction: generateForm.value.instruction,
      count: generateForm.value.count,
      project_id: parseInt(currentProjectId.value)
    })
    generateModalInstance.hide()
    fetchPrompts()
    showModal('Success', 'Prompts generated successfully')
  } catch (error) {
    showModal('Error', 'Failed to generate prompts: ' + (error.response?.data?.error || error.message))
  } finally {
    isGenerating.value = false
  }
}

onMounted(() => {
  fetchPrompts()
  fetchLLMConfigs()
  modalInstance = new Modal(modalRef.value)
  generateModalInstance = new Modal(generateModalRef.value)
  batchTestModalInstance = new Modal(batchTestModalRef.value)
})
</script>
