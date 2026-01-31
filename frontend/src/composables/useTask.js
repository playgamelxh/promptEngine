import { ref } from 'vue'
import axios from 'axios'

// Global state
const currentTaskId = ref(localStorage.getItem('current_llm_task_id') || '')
const isRunning = ref(false)
const taskStatus = ref({
  status: '',
  progress: 0,
  total: 0,
  message: ''
})

export function useTask() {
  const pollStatus = async () => {
    if (!currentTaskId.value) {
      isRunning.value = false
      return
    }

    try {
      const res = await axios.get('/api/llm-test-cases/task/status', {
        params: { task_id: currentTaskId.value }
      })
      taskStatus.value = res.data
      
      if (res.data.status === 'running' || res.data.status === 'pending') {
         isRunning.value = true
         setTimeout(pollStatus, 1000)
      } else {
         isRunning.value = false
         // Task finished (completed, failed, stopped)
         // We keep the status/id so the UI can show the final result
      }
    } catch (error) {
      console.error('Failed to poll status', error)
      isRunning.value = false
      // Don't clear task ID on error, let user see/retry
    }
  }

  const startTask = (id) => {
    currentTaskId.value = id
    localStorage.setItem('current_llm_task_id', id)
    isRunning.value = true
    pollStatus()
  }

  const stopRunningTask = async () => {
    if (!currentTaskId.value) return
    try {
      await axios.post('/api/llm-test-cases/task/stop', {
        task_id: currentTaskId.value
      })
      // Polling will update status to 'stopped'
    } catch (error) {
      console.error('Failed to stop task', error)
      throw error
    }
  }

  const clearTask = () => {
    currentTaskId.value = ''
    localStorage.removeItem('current_llm_task_id')
    taskStatus.value = { status: '', progress: 0, total: 0, message: '' }
    isRunning.value = false
  }
  
  const resumeTask = () => {
    if (currentTaskId.value) {
      pollStatus()
    }
  }

  return {
    currentTaskId,
    isRunning,
    taskStatus,
    startTask,
    stopRunningTask,
    clearTask,
    resumeTask
  }
}
