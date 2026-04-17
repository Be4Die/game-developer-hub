import { createRouter, createWebHistory } from "vue-router";
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
import ModeratorRoles from "../views/ModeratorRoles.vue";
import TicketDetail from "../views/TicketDetail.vue";
import Tickethistory from "../views/Tickethistory.vue";
import Settings from "../views/Settings.vue";

const routes = [
  { path: "/", redirect: "/projects" },
  { path: "/projects", component: ProjectsList },
  {
    path: "/projects/:id",
    component: ProjectWorkspace,
    props: true,
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
  { path: "/nodes", name: "nodes", component: NodesList },
  {
    path: "/nodes/:nodeId",
    name: "node-detail",
    component: NodeDetail,
    props: true,
  },
  { path: "/moderator/tickets", component: ModeratorTickets },
  { path: "/moderator/tickets/:id", component: TicketDetail },
  { path: "/moderator/history", component: Tickethistory },
  { path: "/moderator/roles", component: ModeratorRoles },
  { path: "/settings", name: "settings", component: Settings },
];

export default createRouter({
  history: createWebHistory(),
  routes,
});
