import { ref } from 'vue'

const currentProjectId = ref(localStorage.getItem('currentProjectId') || '')

export function useProject() {
  const setProject = (id) => {
    currentProjectId.value = id
    if (id) {
      localStorage.setItem('currentProjectId', id)
    } else {
      localStorage.removeItem('currentProjectId')
    }
  }

  return {
    currentProjectId,
    setProject
  }
}
