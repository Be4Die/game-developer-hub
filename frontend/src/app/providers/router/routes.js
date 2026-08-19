import { isAuthenticated } from '@/entities/user';
import { LoginPage } from '@/pages/login';
import { ProjectsListPage } from '@/pages/projects-list';
import { ProjectWorkspacePage } from '@/pages/project-workspace';
import { ProjectDraftPage } from '@/pages/project-draft';
import { ProjectStatsPage } from '@/pages/project-stats';
import { ProjectPublishedPage } from '@/pages/project-published';
import { ServersLayoutPage } from '@/pages/servers-layout';
import { ServersOverviewPage } from '@/pages/servers-overview';
import { ServerBuildsPage } from '@/pages/server-builds';
import { ServerInstancesPage } from '@/pages/server-instances';
import { InstanceDetailPage } from '@/pages/instance-detail';
import { NodesListPage } from '@/pages/nodes-list';
import { NodeDetailPage } from '@/pages/node-detail';
import { ModeratorDashboardPage } from '@/pages/moderator-dashboard';
import { ModeratorTicketsPage } from '@/pages/moderator-tickets';
import { TicketDetailPage } from '@/pages/ticket-detail';
import { TicketHistoryPage } from '@/pages/ticket-history';
import { ChatViewPage } from '@/pages/chat-view';
import { SettingsPage } from '@/pages/settings';
import { AdminDashboardPage } from '@/pages/admin-dashboard';

export const routes = [
  {
    path: '/',
    redirect: () => {
      if (!isAuthenticated()) return '/login';
      const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
      if (user?.role === 'USER_ROLE_ADMIN') return '/admin/dashboard';
      if (user?.role === 'USER_ROLE_MODERATOR') return '/moderator';
      return '/projects';
    },
  },
  {
    path: '/login',
    name: 'login',
    component: LoginPage,
    meta: { guest: true },
  },
  {
    path: '/projects',
    name: 'projects',
    component: ProjectsListPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/projects/:id',
    component: ProjectWorkspacePage,
    props: true,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: (to) => `/projects/${to.params.id}/draft` },
      { path: 'draft', name: 'draft', component: ProjectDraftPage },
      { path: 'stats', name: 'stats', component: ProjectStatsPage },
      { path: 'published', name: 'published', component: ProjectPublishedPage },
      {
        path: 'servers',
        component: ServersLayoutPage,
        props: (route) => ({ gameId: route.params.id }),
        children: [
          {
            path: '',
            name: 'servers',
            component: ServersOverviewPage,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: 'builds',
            name: 'server-builds',
            component: ServerBuildsPage,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: 'instances',
            name: 'server-instances',
            component: ServerInstancesPage,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: 'instances/:instanceId',
            name: 'instance-detail',
            component: InstanceDetailPage,
            props: (route) => ({
              gameId: route.params.id,
              instanceId: route.params.instanceId,
            }),
          },
        ],
      },
    ],
  },
  {
    path: '/nodes',
    name: 'nodes',
    component: NodesListPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/nodes/:nodeId',
    name: 'node-detail',
    component: NodeDetailPage,
    props: true,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator',
    name: 'moderator',
    component: ModeratorDashboardPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/tickets',
    name: 'moderator-tickets',
    component: ModeratorTicketsPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/tickets/:id',
    name: 'ticket-detail',
    component: TicketDetailPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/history',
    name: 'ticket-history',
    component: TicketHistoryPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/chat/:id',
    name: 'chat-view',
    component: ChatViewPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/admin/dashboard',
    name: 'admin-dashboard',
    component: AdminDashboardPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
];
