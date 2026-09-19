import { reactive } from 'vue';

export interface ToastState {
  show: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info' | string;
}

export const toast = reactive<ToastState>({
  show: false,
  message: '',
  type: 'success',
});

let toastTimer: ReturnType<typeof setTimeout> | null = null;

export function showToast(message: string, type: ToastState['type'] = 'success'): void {
  if (toastTimer) {
    clearTimeout(toastTimer);
  }
  toast.message = message;
  toast.type = type;
  toast.show = true;
  toastTimer = setTimeout(() => {
    toast.show = false;
    toastTimer = null;
  }, 3000);
}
