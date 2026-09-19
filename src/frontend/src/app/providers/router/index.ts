import { createRouter, createWebHistory } from 'vue-router';
import { isAuthenticated } from '@/entities/user';
import { routes } from './routes';

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authed = isAuthenticated();
  if (to.meta.requiresAuth && !authed) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  if (to.meta.guest && authed) {
    const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
    if (user?.role === 'USER_ROLE_ADMIN' || user?.role === 3) return { path: '/catalog' };
    if (user?.role === 'USER_ROLE_MODERATOR' || user?.role === 2)
      return { path: '/moderator/queue' };
    return { path: '/projects' };
  }
  if (to.meta.requiresAdmin) {
    const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
    if (user?.role !== 'USER_ROLE_ADMIN' && user?.role !== 3) {
      return { path: '/projects' };
    }
  }
  if (to.meta.requiresStaff) {
    const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
    const r = user?.role;
    const isStaff =
      r === 'USER_ROLE_ADMIN' ||
      r === 'admin' ||
      r === 3 ||
      r === 'USER_ROLE_MODERATOR' ||
      r === 'moderator' ||
      r === 2;
    if (!isStaff) {
      return { path: '/projects' };
    }
  }
});

export default router;
