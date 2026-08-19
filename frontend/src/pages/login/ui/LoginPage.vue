<template>
  <div class="auth-page">
    <div class="theme-toggle-wrap">
      <ThemeToggle />
    </div>
    <div class="auth-brand">
      <LogoIcon :size="48" :textSize="28" :showSub="true" :glow="true" />
    </div>

    <AuthForm />
  </div>
</template>

<script setup>
import { LogoIcon } from '@/shared/ui';
import { ThemeToggle } from '@/features/theme-switcher';
import { AuthForm } from '@/features/auth-by-email';
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-app);
  padding: 24px;
  gap: 24px;
  position: relative;
}

.theme-toggle-wrap {
  position: absolute;
  top: 20px;
  right: 20px;
}

.auth-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
</style>
