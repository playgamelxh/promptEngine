<template>
  <div class="d-flex flex-column min-vh-100">
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
      <div class="container-fluid">
        <router-link class="navbar-brand" to="/">CodeAgent</router-link>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav me-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/" active-class="active">Home</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/llm-configs" active-class="active">LLM Configs</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/projects" active-class="active">Projects</router-link>
            </li>
            <li class="nav-item" v-if="currentProjectId">
              <router-link class="nav-link" to="/prompts" active-class="active">Prompts</router-link>
            </li>
            <li class="nav-item" v-if="currentProjectId">
              <router-link class="nav-link" to="/test-cases" active-class="active">Test Cases</router-link>
            </li>
            <li class="nav-item" v-if="currentProjectId">
              <router-link class="nav-link" to="/llm-test-cases" active-class="active">LLM Test Cases</router-link>
            </li>
          </ul>
          <div class="d-flex align-items-center">
             <span v-if="currentProjectName" class="text-light me-3">Project: <strong>{{ currentProjectName }}</strong></span>
             <span v-else class="text-warning me-3">No Project Selected</span>
          </div>
        </div>
      </div>
    </nav>

    <div class="container mt-4 flex-grow-1">
      <div v-if="!currentProjectId && $route.path !== '/projects' && $route.path !== '/llm-configs' && $route.path !== '/'" class="alert alert-warning">
        Please select a project first in the <router-link to="/projects">Projects</router-link> page.
      </div>
      <router-view />
    </div>

    <footer class="bg-light text-center text-lg-start mt-auto py-3">
      <div class="container text-center text-muted">
        © 2026 CodeAgent
      </div>
    </footer>

    <!-- Global Info Modal -->
    <div class="modal fade" ref="globalModalRef" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ modalTitle }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <p>{{ modalBody }}</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Global Confirm Modal -->
    <div class="modal fade" ref="confirmModalRef" tabindex="-1" aria-hidden="true" data-bs-backdrop="static" data-bs-keyboard="false">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ confirmTitle }}</h5>
            <button type="button" class="btn-close" @click="handleConfirm(false)" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <p>{{ confirmBody }}</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="handleConfirm(false)">Cancel</button>
            <button type="button" class="btn btn-danger" @click="handleConfirm(true)">Confirm</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Global Task Status Widget -->
    <div v-if="currentTaskId" class="position-fixed bottom-0 end-0 p-3" style="z-index: 1050; width: 350px;">
      <div class="card shadow-sm border-info">
        <div class="card-header bg-info text-white d-flex justify-content-between align-items-center py-2">
          <h6 class="mb-0 small"><i class="bi bi-cpu me-1"></i>Task: {{ taskStatus.status || 'Initializing...' }}</h6>
          <button v-if="!isRunning" type="button" class="btn-close btn-close-white small" @click="clearTask" aria-label="Close"></button>
        </div>
        <div class="card-body py-2">
           <div class="d-flex justify-content-between mb-1 small">
             <span class="text-truncate" style="max-width: 200px;">{{ taskStatus.message }}</span>
             <span>{{ taskStatus.progress }} / {{ taskStatus.total }}</span>
           </div>
           <div class="progress" style="height: 8px;">
              <div class="progress-bar" 
                   :class="{ 
                     'progress-bar-striped progress-bar-animated': isRunning, 
                     'bg-success': taskStatus.status === 'completed', 
                     'bg-danger': taskStatus.status === 'failed', 
                     'bg-secondary': taskStatus.status === 'stopped' 
                   }" 
                   role="progressbar" 
                   :style="{ width: (taskStatus.total > 0 ? (taskStatus.progress / taskStatus.total * 100) : 0) + '%' }">
              </div>
           </div>
           <div class="mt-2 text-end" v-if="isRunning">
              <button class="btn btn-xs btn-outline-danger py-0 px-2" style="font-size: 0.8rem;" @click="handleStopTask">Stop Task</button>
           </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, provide } from 'vue'
import { useProject } from './composables/useProject'
import { useTask } from './composables/useTask'
import axios from 'axios'
import { Modal } from 'bootstrap'

const { currentProjectId } = useProject()
const { currentTaskId, isRunning, taskStatus, stopRunningTask, clearTask, resumeTask } = useTask()
const currentProjectName = ref('')

// Global Modal Logic
const globalModalRef = ref(null)
const modalTitle = ref('')
const modalBody = ref('')
let globalModalInstance = null

const showGlobalModal = (title, message) => {
  modalTitle.value = title
  modalBody.value = message
  if (globalModalInstance) {
    globalModalInstance.show()
  }
}

provide('showGlobalModal', showGlobalModal)

// Global Confirm Modal Logic
const confirmModalRef = ref(null)
const confirmTitle = ref('')
const confirmBody = ref('')
let confirmModalInstance = null
let confirmResolve = null

const showGlobalConfirm = (title, message) => {
  confirmTitle.value = title
  confirmBody.value = message
  if (confirmModalInstance) {
    confirmModalInstance.show()
  }
  return new Promise((resolve) => {
    confirmResolve = resolve
  })
}

const handleConfirm = (result) => {
  if (confirmResolve) {
    confirmResolve(result)
    confirmResolve = null
  }
  if (confirmModalInstance) {
    confirmModalInstance.hide()
  }
}

provide('showGlobalConfirm', showGlobalConfirm)

const handleStopTask = async () => {
  if (await showGlobalConfirm('Stop Task', 'Are you sure you want to stop the current task?')) {
    try {
      await stopRunningTask()
    } catch (e) {
      showGlobalModal('Error', 'Failed to stop task')
    }
  }
}

const fetchProjectName = async () => {
  if (!currentProjectId.value) {
    currentProjectName.value = ''
    return
  }
  try {
    const res = await axios.get('/api/projects', { params: { page: 1, page_size: 1000 } }) // Fetch enough to find the project
    const items = res.data.items || (Array.isArray(res.data) ? res.data : [])
    const project = items.find(p => p.id == currentProjectId.value)
    if (project) {
      currentProjectName.value = project.name
    } else {
      currentProjectName.value = 'Unknown'
    }
  } catch (e) {
    console.error(e)
  }
}

watch(currentProjectId, () => {
  fetchProjectName()
})

onMounted(() => {
  if (globalModalRef.value) {
    globalModalInstance = new Modal(globalModalRef.value)
  }
  if (confirmModalRef.value) {
    confirmModalInstance = new Modal(confirmModalRef.value)
  }
  fetchProjectName()
  resumeTask()
})
</script>

<style>
/* Custom active link style if needed, though bootstrap handles it via .active class */
.nav-link.active {
  font-weight: bold;
  color: #fff !important;
}
</style>
