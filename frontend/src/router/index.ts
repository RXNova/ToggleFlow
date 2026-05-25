import { createRouter, createWebHistory } from 'vue-router'

// Vue Router is like Angular Router — maps URL paths to view components.
// createWebHistory() uses the HTML5 History API (no # in URLs), same as Angular default.
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/components/layout/AppLayout.vue'),
      children: [
        {
          path: '',
          redirect: '/flags',
        },
        {
          path: 'flags',
          name: 'flags',
          component: () => import('@/views/FlagsView.vue'),
        },
        {
          path: 'environments',
          name: 'environments',
          component: () => import('@/views/EnvironmentsView.vue'),
        },
        {
          path: 'audit',
          name: 'audit',
          component: () => import('@/views/AuditView.vue'),
        },
      ],
    },
  ],
})

export default router
