import { isAuthenticated } from '@/entities/user';
import { LoginPage } from '@/pages/login';
import { ProjectsListPage } from '@/pages/projects-list';
import { ProjectWorkspacePage } from '@/pages/project-workspace';
import { ProjectDraftPage } from '@/pages/project-draft';
import { ProjectSandboxPage } from '@/pages/project-sandbox';
import { ProjectStatsPage } from '@/pages/project-stats';
import { ProjectPublishedPage } from '@/pages/project-published';
import { ServersLayoutPage } from '@/pages/servers-layout';
import { ServersOverviewPage } from '@/pages/servers-overview';
import { ServerBuildsPage } from '@/pages/server-builds';
import { ServerInstancesPage } from '@/pages/server-instances';
import { InstanceDetailPage } from '@/pages/instance-detail';
import { NodesListPage } from '@/pages/nodes-list';
import { NodeDetailPage } from '@/pages/node-detail';
import { ModerationQueuePage } from '@/pages/moderation-queue';
import { ModeratorChatsPage } from '@/pages/moderator-chats';
import { ModerationArchivePage } from '@/pages/moderation-archive';
import { ModerationProjectPage } from '@/pages/moderation-project';
import { ModerationRejectPage } from '@/pages/moderation-reject';
import { ModerationSnapshotPage } from '@/pages/moderation-snapshot';
import { ProfilePage } from '@/pages/profile';
import { AdminDevelopersPage } from '@/pages/admin-developers';
import { AdminModeratorsPage } from '@/pages/admin-moderators';
import { AdminModeratorDetailPage } from '@/pages/admin-moderator-detail';
import { ProjectAccessPage } from '@/pages/project-access';
import { ProjectPurchasesPage } from '@/pages/project-purchases';
import { CatalogPage } from '@/pages/catalog';
import { DocsPage } from '@/pages/docs';

export const routes = [
  {
    path: '/',
    redirect: () => {
      if (!isAuthenticated()) return '/login';
      const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
      if (user?.role === 'USER_ROLE_ADMIN' || user?.role === 3) return '/catalog';
      if (user?.role === 'USER_ROLE_MODERATOR' || user?.role === 2) return '/moderator/queue';
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
    path: '/access',
    redirect: '/profile',
  },
  {
    path: '/projects/:id',
    component: ProjectWorkspacePage,
    props: true,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: (to) => `/projects/${to.params.id}/draft` },
      { path: 'draft', name: 'draft', component: ProjectDraftPage },
      { path: 'sandbox', name: 'sandbox', component: ProjectSandboxPage },
      { path: 'purchases', name: 'project-purchases', component: ProjectPurchasesPage },
      { path: 'stats', name: 'stats', component: ProjectStatsPage },
      { path: 'published', name: 'published', component: ProjectPublishedPage },
      { path: 'access', name: 'project-access', component: ProjectAccessPage },
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
    redirect: '/moderator/queue',
  },
  {
    path: '/catalog',
    name: 'catalog',
    component: CatalogPage,
    meta: { requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/admin/catalog',
    redirect: '/catalog',
  },
  {
    path: '/moderator/catalog',
    redirect: '/catalog',
  },
  {
    path: '/moderator/queue',
    name: 'moderation-queue',
    component: ModerationQueuePage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/chats',
    name: 'moderator-chats',
    component: ModeratorChatsPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/journal',
    name: 'moderation-journal',
    component: ModerationArchivePage,
    meta: { requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/moderator/platform-servers',
    redirect: '/moderator/queue?type=server',
  },
  {
    path: '/admin/platform-servers',
    redirect: '/moderator/queue?type=server',
  },
  {
    path: '/moderator/archive',
    redirect: '/moderator/journal',
  },
  {
    path: '/admin/journal',
    redirect: '/moderator/journal',
  },
  {
    path: '/moderator/projects/:projectId',
    name: 'moderation-project',
    component: ModerationProjectPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/projects/:projectId/reject',
    name: 'moderation-reject',
    component: ModerationRejectPage,
    meta: { requiresAuth: true },
  },
  {
    path: '/moderator/journal/:requestId',
    name: 'moderation-journal-detail',
    component: ModerationSnapshotPage,
    meta: { requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/moderator/snapshots/:requestId',
    redirect: (to) => `/moderator/journal/${to.params.requestId}`,
  },
  {
    path: '/moderator/archive/:requestId',
    redirect: (to) => `/moderator/journal/${to.params.requestId}`,
  },
  {
    path: '/profile',
    name: 'profile',
    component: ProfilePage,
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    redirect: '/profile',
  },
  {
    path: '/admin',
    redirect: '/catalog',
  },
  {
    path: '/admin/dashboard',
    redirect: '/catalog',
  },
  {
    path: '/admin/developers',
    name: 'admin-developers',
    component: AdminDevelopersPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/admin/moderators',
    name: 'admin-moderators',
    component: AdminModeratorsPage,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/admin/moderators/:id',
    name: 'admin-moderator-detail',
    component: AdminModeratorDetailPage,
    props: true,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/docs',
    redirect: '/docs/getting-started/overview',
  },
  {
    path: '/docs/:section',
    name: 'docs-section',
    component: DocsPage,
  },
  {
    path: '/docs/:section/:article',
    name: 'docs-article',
    component: DocsPage,
  },
  {
    path: '/rules',
    redirect: '/docs/rules/rules-catalog',
  },
];
