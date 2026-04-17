<template>
  <div class="servers-layout">
    <div class="servers-subnav">
      <router-link :to="`/projects/${gameId}/servers`" class="subnav-btn" exact-active-class="active">
        <LayoutDashboard class="icon-sm" /> Обзор
      </router-link>
      <router-link :to="`/projects/${gameId}/servers/builds`" class="subnav-btn" active-class="active">
        <Package class="icon-sm" /> Билды
      </router-link>
      <router-link :to="`/projects/${gameId}/servers/instances`" class="subnav-btn" active-class="active">
        <Play class="icon-sm" /> Инстансы
      </router-link>
    </div>
    <div class="servers-content scrollable">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { LayoutDashboard, Package, Play } from 'lucide-vue-next'
defineProps({ gameId: { type: [String, Number], required: true } })
</script>

<style scoped>
.servers-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.servers-subnav {
  display: flex;
  gap: 4px;
  padding: 0 32px 16px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 0;
  background: var(--bg-card);
}
.subnav-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  text-decoration: none;
  transition: 0.15s;
}
.subnav-btn:hover { background: var(--bg-hover); color: var(--text-main); }
.subnav-btn.active { background: #EFF6FF; color: var(--primary); font-weight: 600; }
[data-theme="dark"] .subnav-btn.active { background: #1e3a5f; }

.servers-content {
  flex: 1;
  padding: 24px 32px;
  overflow-y: auto;
}
.scrollable { overflow-y: auto; }
</style>
