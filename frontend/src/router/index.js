import { createRouter, createWebHistory } from "vue-router";
import { isAuthenticated } from "../store/auth";
import Login from "../views/Login.vue";
import ProjectsList from "../views/ProjectsList.vue";
import ProjectWorkspace from "../views/ProjectWorkspace.vue";
import DraftTab from "../views/tabs/DraftTab.vue";
import StatsTab from "../views/tabs/StatsTab.vue";
import PublishedTab from "../views/tabs/PublishedTab.vue";
import ServersLayout from "../views/servers/ServersLayout.vue";
import ServersOverview from "../views/servers/ServersOverview.vue";
import ServerBuilds from "../views/servers/ServerBuilds.vue";
import ServerInstances from "../views/servers/ServerInstances.vue";
import InstanceDetail from "../views/servers/InstanceDetail.vue";
import NodesList from "../views/nodes/NodesList.vue";
import NodeDetail from "../views/nodes/NodeDetail.vue";
import ModeratorTickets from "../views/ModeratorTickets.vue";
import TicketDetail from "../views/TicketDetail.vue";
import Tickethistory from "../views/Tickethistory.vue";
import Settings from "../views/Settings.vue";
import AdminDashboard from "../views/AdminDashboard.vue";

const routes = [
  {
    path: "/",
    redirect: () => {
      if (!isAuthenticated()) return "/login";
      const user = JSON.parse(localStorage.getItem("gdh_user") || "null");
      if (user?.role === "USER_ROLE_ADMIN") return "/admin/dashboard";
      if (user?.role === "USER_ROLE_MODERATOR") return "/moderator/tickets";
      return "/projects";
    },
  },
  { path: "/login", component: Login, meta: { guest: true } },
  { path: "/projects", component: ProjectsList, meta: { requiresAuth: true } },
  {
    path: "/projects/:id",
    component: ProjectWorkspace,
    props: true,
    meta: { requiresAuth: true },
    children: [
      { path: "", redirect: (to) => `/projects/${to.params.id}/draft` },
      { path: "draft", name: "draft", component: DraftTab },
      { path: "stats", name: "stats", component: StatsTab },
      { path: "published", name: "published", component: PublishedTab },
      {
        path: "servers",
        component: ServersLayout,
        props: (route) => ({ gameId: route.params.id }),
        children: [
          {
            path: "",
            name: "servers",
            component: ServersOverview,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: "builds",
            name: "server-builds",
            component: ServerBuilds,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: "instances",
            name: "server-instances",
            component: ServerInstances,
            props: (route) => ({ gameId: route.params.id }),
          },
          {
            path: "instances/:instanceId",
            name: "instance-detail",
            component: InstanceDetail,
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
    path: "/nodes",
    name: "nodes",
    component: NodesList,
    meta: { requiresAuth: true },
  },
  {
    path: "/nodes/:nodeId",
    name: "node-detail",
    component: NodeDetail,
    props: true,
    meta: { requiresAuth: true },
  },
  {
    path: "/moderator/tickets",
    component: ModeratorTickets,
    meta: { requiresAuth: true },
  },
  {
    path: "/moderator/tickets/:id",
    component: TicketDetail,
    meta: { requiresAuth: true },
  },
  {
    path: "/moderator/history",
    component: Tickethistory,
    meta: { requiresAuth: true },
  },
  {
    path: "/settings",
    name: "settings",
    component: Settings,
    meta: { requiresAuth: true },
  },
  {
    path: "/admin/dashboard",
    name: "admin-dashboard",
    component: AdminDashboard,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authed = isAuthenticated();
  if (to.meta.requiresAuth && !authed) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  if (to.meta.guest && authed) {
    return { path: "/projects" };
  }
  if (to.meta.requiresAdmin) {
    const user = JSON.parse(localStorage.getItem("gdh_user") || "null");
    if (user?.role !== "USER_ROLE_ADMIN") {
      return { path: "/projects" };
    }
  }
});

export default router;
