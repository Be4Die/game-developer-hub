import { createApp } from 'vue';
import { App, router } from './app';
import { i18n } from '@/shared/lib';
import '@/shared/assets/design-system.css';
import '@/shared/assets/main.css';

const app = createApp(App);

app.use(router);
app.use(i18n);

app.mount('#app');
