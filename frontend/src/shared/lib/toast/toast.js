import { reactive } from 'vue';

export const toast = reactive({
  show: false,
  message: '',
  type: 'success',
});

let toastTimer = null;

export function showToast(message, type = 'success') {
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
