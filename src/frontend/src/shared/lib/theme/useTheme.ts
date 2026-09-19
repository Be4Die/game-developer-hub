import { ref, type Ref } from 'vue';

function getInitialTheme(): boolean {
  if (typeof window === 'undefined') return false;
  const saved = localStorage.getItem('theme');
  if (saved === 'dark') return true;
  if (saved === 'light') return false;
  return Boolean(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
}

const isDark = ref(getInitialTheme());

function applyTheme(): void {
  if (typeof document === 'undefined') return;
  if (isDark.value) {
    document.documentElement.setAttribute('data-theme', 'dark');
  } else {
    document.documentElement.removeAttribute('data-theme');
  }
}

// Immediately apply initial theme
if (typeof window !== 'undefined') {
  applyTheme();

  // Listen to OS system theme changes if user hasn't explicitly chosen a theme
  if (window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (!localStorage.getItem('theme')) {
        isDark.value = e.matches;
        applyTheme();
      }
    });
  }
}

export function useTheme(): {
  isDark: Ref<boolean>;
  toggleTheme: () => void;
  setTheme: (value: 'dark' | 'light') => void;
} {
  function toggleTheme(): void {
    isDark.value = !isDark.value;
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light');
    applyTheme();
  }

  function setTheme(value: 'dark' | 'light'): void {
    isDark.value = value === 'dark';
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light');
    applyTheme();
  }

  return { isDark, toggleTheme, setTheme };
}
