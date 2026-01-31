import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/llm-configs',
      name: 'llm-configs',
      component: () => import('../views/LLMConfigView.vue')
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('../views/ProjectView.vue')
    },
    {
      path: '/prompts',
      name: 'prompts',
      component: () => import('../views/PromptView.vue')
    },
    {
      path: '/test-cases',
      name: 'test-cases',
      component: () => import('../views/TestCaseView.vue')
    },
    {
      path: '/llm-test-cases',
      name: 'llm-test-cases',
      component: () => import('../views/LLMTestCaseView.vue')
    }
  ]
})

export default router
