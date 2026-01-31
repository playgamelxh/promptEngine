<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Projects</h2>
      <div>
        <button class="btn btn-danger me-2" v-if="selectedIds.length > 0" @click="batchDelete">Delete Selected ({{ selectedIds.length }})</button>
        <button class="btn btn-primary" @click="openDialog()">Add Project</button>
      </div>
    </div>

    <table class="table table-striped table-hover">
      <thead class="table-dark">
        <tr>
          <th scope="col" style="width: 40px;">
            <input type="checkbox" class="form-check-input" :checked="isAllSelected" @change="toggleSelectAll">
          </th>
          <th scope="col">ID</th>
          <th scope="col">Name</th>
          <th scope="col" style="width: 40%;">Description</th>
          <th scope="col">Tags</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="project in projects" :key="project.id" :class="{'table-success': project.id == currentProjectId}">
          <td>
            <input type="checkbox" class="form-check-input" :value="project.id" v-model="selectedIds">
          </td>
          <td>{{ project.id }}</td>
          <td>
            {{ project.name }}
            <span v-if="project.id == currentProjectId" class="badge bg-success ms-2">Selected</span>
          </td>
          <td>
            <div class="text-wrap" :title="project.description" style="display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; max-height: 4.5em;">{{ project.description }}</div>
          </td>
          <td>
            <span v-for="tag in project.tags ? project.tags.split(',') : []" :key="tag" class="badge bg-secondary me-1">{{ tag.trim() }}</span>
          </td>
          <td>
            <button v-if="project.id != currentProjectId" class="btn btn-sm btn-outline-success me-2" @click="selectProject(project.id)">Select</button>
            <button class="btn btn-sm btn-outline-primary me-2" @click="openDialog(project)">Edit</button>
            <button class="btn btn-sm btn-outline-danger" @click="deleteProject(project.id)">Delete</button>
          </td>
        </tr>
        <tr v-if="projects.length === 0">
          <td colspan="6" class="text-center">No projects found.</td>
        </tr>
      </tbody>
    </table>

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
    <div class="modal fade" id="projectModal" tabindex="-1" aria-hidden="true" ref="modalRef">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? 'Edit Project' : 'Add Project' }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveProject">
              <div class="mb-3">
                <label class="form-label">Name</label>
                <input type="text" class="form-control" v-model="form.name" required>
              </div>
              <div class="mb-3">
                <label class="form-label">Description</label>
                <textarea class="form-control" v-model="form.description" rows="3"></textarea>
              </div>
              <div class="mb-3">
                <label class="form-label">Tags</label>
                <input type="text" class="form-control" v-model="form.tags" placeholder="Comma separated">
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" @click="saveProject">Confirm</button>
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
import { useProject } from '../composables/useProject'

const showModal = inject('showGlobalModal')
const showConfirm = inject('showGlobalConfirm')

const projects = ref([])
const selectedIds = ref([])
const currentPage = ref(1)
const pageSize = ref(30)
const totalItems = ref(0)
const form = ref({
  id: null,
  name: '',
  description: '',
  tags: ''
})
const modalRef = ref(null)
let modalInstance = null
const { currentProjectId, setProject: setGlobalProject } = useProject()

const isAllSelected = computed(() => {
  return projects.value.length > 0 && selectedIds.value.length === projects.value.length
})

const totalPages = computed(() => {
  return Math.ceil(totalItems.value / pageSize.value)
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = projects.value.map(p => p.id)
  }
}

const batchDelete = async () => {
  if (!await showConfirm('Confirm', `Are you sure you want to delete ${selectedIds.value.length} projects?`)) return
  try {
    await axios.delete('/api/projects/batch', { data: { ids: selectedIds.value } })
    selectedIds.value = []
    fetchProjects()
  } catch (error) {
    showModal('Error', 'Failed to delete selected projects')
  }
}

const fetchProjects = async () => {
  try {
    const res = await axios.get('/api/projects', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value
      }
    })
    if (res.data && res.data.items) {
      projects.value = res.data.items
      totalItems.value = res.data.total
    } else {
      // Fallback for non-paginated response or empty
      projects.value = Array.isArray(res.data) ? res.data : []
      totalItems.value = projects.value.length
    }
    selectedIds.value = []
  } catch (error) {
    console.error(error)
    showModal('Error', 'Failed to fetch projects')
  }
}

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchProjects()
}

const openDialog = (row) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = { id: null, name: '', description: '', tags: '' }
  }
  modalInstance.show()
}

const saveProject = async () => {
  try {
    if (form.value.id) {
      await axios.put(`/api/projects/${form.value.id}`, form.value)
    } else {
      await axios.post('/api/projects', form.value)
    }
    modalInstance.hide()
    fetchProjects()
  } catch (error) {
    showModal('Error', 'Failed to save')
  }
}

const deleteProject = async (id) => {
  if (!await showConfirm('Confirm', 'Are you sure you want to delete this project?')) return
  try {
    await axios.delete(`/api/projects/${id}`)
    fetchProjects()
    // If deleted project was selected, clear selection
    if (localStorage.getItem('currentProjectId') == id) {
      setGlobalProject('')
    }
  } catch (error) {
    showModal('Error', 'Failed to delete')
  }
}

const selectProject = (id) => {
  setGlobalProject(id)
  showModal('Success', 'Project selected successfully')
}

onMounted(() => {
  fetchProjects()
  modalInstance = new Modal(modalRef.value)
})
</script>
