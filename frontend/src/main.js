import { createApp } from 'vue';
import { App, router } from './app';
import '@/shared/assets/design-system.css';
import '@/shared/assets/main.css';

const app = createApp(App);

app.use(router);

app.mount('#app');
