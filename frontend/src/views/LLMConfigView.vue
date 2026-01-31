<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>LLM Configs</h2>
      <div>
        <button class="btn btn-danger me-2" v-if="selectedIds.length > 0" @click="batchDelete">Delete Selected ({{ selectedIds.length }})</button>
        <button class="btn btn-primary" @click="openDialog()">Add LLM Config</button>
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
            <th scope="col">Model</th>
            <th scope="col">Base URL</th>
            <th scope="col">Temperature</th>
            <th scope="col">Tags</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="config in configs" :key="config.id">
            <td>
              <input type="checkbox" class="form-check-input" :value="config.id" v-model="selectedIds">
            </td>
            <td>{{ config.id }}</td>
            <td>
              {{ config.name }}
              <span v-if="config.is_default" class="badge bg-success ms-2">Default</span>
            </td>
            <td>{{ config.model_name }}</td>
            <td>
              <div class="text-truncate" :title="config.base_url" style="max-width: 200px;">{{ config.base_url }}</div>
            </td>
            <td>{{ config.temperature }}</td>
            <td>
              <span v-for="tag in config.tags ? config.tags.split(',') : []" :key="tag" class="badge bg-secondary me-1">{{ tag.trim() }}</span>
            </td>
            <td>
              <button class="btn btn-sm btn-outline-success me-2" v-if="!config.is_default" @click="setDefault(config)">Set Default</button>
              <button class="btn btn-sm btn-outline-primary me-2" @click="openDialog(config)">Edit</button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteConfig(config.id)">Delete</button>
            </td>
          </tr>
          <tr v-if="configs.length === 0">
            <td colspan="8" class="text-center">No configurations found.</td>
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
    <div class="modal fade" id="configModal" tabindex="-1" aria-hidden="true" ref="modalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? 'Edit Config' : 'Add Config' }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Load Preset</label>
              <select class="form-select" @change="loadPreset($event)">
                <option value="">Select a preset...</option>
                <option value="ollama-native">Ollama (Native API /api/generate)</option>
                <option value="ollama-r1">Ollama (OpenAI Compatible /v1)</option>
                <option value="deepseek">DeepSeek (Official API)</option>
                <option value="openai">OpenAI (GPT-4o)</option>
              </select>
            </div>
            <hr>
            <form @submit.prevent="saveConfig">
              <div class="mb-3">
                <label class="form-label">Name</label>
                <input type="text" class="form-control" v-model="form.name" required>
              </div>
              <div class="mb-3">
                <label class="form-label">API Key</label>
                <input type="password" class="form-control" v-model="form.api_key">
              </div>
              <div class="mb-3">
                <label class="form-label">Base URL</label>
                <input type="text" class="form-control" v-model="form.base_url">
              </div>
              <div class="mb-3">
                <label class="form-label">Model Name</label>
                <input type="text" class="form-control" v-model="form.model_name">
              </div>
              <div class="mb-3">
                <label class="form-label">Temperature (0.0 - 2.0)</label>
                <input type="number" class="form-control" v-model.number="form.temperature" min="0" max="2" step="0.1">
              </div>
              <div class="mb-3">
                <label class="form-label">Tags</label>
                <input type="text" class="form-control" v-model="form.tags" placeholder="Comma separated">
              </div>
              <div class="form-check mb-3">
                <input class="form-check-input" type="checkbox" v-model="form.is_default" id="isDefaultCheck">
                <label class="form-check-label" for="isDefaultCheck">
                  Set as Default Config
                </label>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="saveConfig">Confirm</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, inject } from 'vue'
import axios from 'axios'
import { Modal } from 'bootstrap'

const showModal = inject('showGlobalModal')
const showConfirm = inject('showGlobalConfirm')

const configs = ref([])
const selectedIds = ref([])
const currentPage = ref(1)
const pageSize = ref(30)
const totalItems = ref(0)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))

const form = ref({
  id: null,
  name: '',
  api_key: '',
  base_url: '',
  model_name: '',
  temperature: 0.7,
  tags: '',
  is_default: false
})
const modalRef = ref(null)
let modalInstance = null

const isAllSelected = computed(() => {
  return configs.value.length > 0 && selectedIds.value.length === configs.value.length
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = configs.value.map(c => c.id)
  }
}

const batchDelete = async () => {
  if (!await showConfirm('Confirm', `Are you sure you want to delete ${selectedIds.value.length} configs?`)) return
  try {
    await axios.delete('/api/llm-configs/batch', { data: { ids: selectedIds.value } })
    selectedIds.value = []
    fetchConfigs()
  } catch (error) {
    showModal('Error', 'Failed to delete selected configs')
  }
}

const fetchConfigs = async () => {
  try {
    const res = await axios.get('/api/llm-configs', {
      params: { page: currentPage.value, page_size: pageSize.value }
    })
    if (res.data && res.data.items !== undefined) {
      configs.value = res.data.items || []
      totalItems.value = res.data.total || 0
    } else if (Array.isArray(res.data)) {
      configs.value = res.data
      totalItems.value = res.data.length
    } else {
      configs.value = []
      totalItems.value = 0
    }
    selectedIds.value = []
  } catch (error) {
    showModal('Error', 'Failed to fetch configs')
  }
}

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchConfigs()
}

const openDialog = (row) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = { id: null, name: '', api_key: '', base_url: '', model_name: '', temperature: 0.7, tags: '', is_default: false }
  }
  modalInstance.show()
}

const loadPreset = (event) => {
  const preset = event.target.value
  if (!preset) return

  if (preset === 'ollama-native') {
    form.value.name = 'Ollama Native'
    form.value.base_url = 'http://host.docker.internal:11434/api/generate'
    form.value.model_name = 'deepseek-r1:7b'
    form.value.api_key = 'ollama'
    form.value.temperature = 0.7
    form.value.tags = 'ollama,local,native'
  } else if (preset === 'ollama-r1') {
    form.value.name = 'Ollama DeepSeek R1'
    form.value.base_url = 'http://host.docker.internal:11434/v1'
    form.value.model_name = 'deepseek-r1:7b'
    form.value.api_key = 'ollama'
    form.value.temperature = 0.6
    form.value.tags = 'ollama,local,deepseek'
  } else if (preset === 'deepseek') {
    form.value.name = 'DeepSeek API'
    form.value.base_url = 'https://api.deepseek.com'
    form.value.model_name = 'deepseek-chat'
    form.value.temperature = 1.0
    form.value.tags = 'deepseek,api'
  } else if (preset === 'openai') {
    form.value.name = 'OpenAI GPT-4o'
    form.value.base_url = 'https://api.openai.com/v1'
    form.value.model_name = 'gpt-4o'
    form.value.temperature = 0.7
    form.value.tags = 'openai,gpt-4o'
  }
  // Reset dropdown
  event.target.value = ''
}

const setDefault = async (config) => {
  try {
    const updatedConfig = { ...config, is_default: true }
    await axios.put(`/api/llm-configs/${config.id}`, updatedConfig)
    fetchConfigs()
  } catch (error) {
    showModal('Error', 'Failed to set default: ' + (error.response?.data?.error || error.message))
  }
}

const saveConfig = async () => {
  try {
    if (form.value.id) {
      await axios.put(`/api/llm-configs/${form.value.id}`, form.value)
    } else {
      await axios.post('/api/llm-configs', form.value)
    }
    modalInstance.hide()
    fetchConfigs()
  } catch (error) {
    showModal('Error', 'Failed to save')
  }
}

const deleteConfig = async (id) => {
  if (!await showConfirm('Confirm', 'Are you sure you want to delete this config?')) return
  try {
    await axios.delete(`/api/llm-configs/${id}`)
    fetchConfigs()
  } catch (error) {
    showModal('Error', 'Failed to delete')
  }
}

onMounted(() => {
  fetchConfigs()
  modalInstance = new Modal(modalRef.value)
})
</script>
