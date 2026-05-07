import { ref, onMounted } from "vue";

const isDark = ref(false);

export function useTheme() {
    onMounted(() => {
        const saved = localStorage.getItem("theme");
        isDark.value = saved === "dark" || (!saved && window.matchMedia("(prefers-color-scheme: dark)").matches);
        applyTheme();
    });

    function applyTheme() {
        if (isDark.value) {
            document.documentElement.setAttribute("data-theme", "dark");
        } else {
            document.documentElement.removeAttribute("data-theme");
        }
    }

    function toggleTheme() {
        isDark.value = !isDark.value;
        localStorage.setItem("theme", isDark.value ? "dark" : "light");
        applyTheme();
    }

    function setTheme(value) {
        isDark.value = value === "dark";
        localStorage.setItem("theme", isDark.value ? "dark" : "light");
        applyTheme();
    }

    return { isDark, toggleTheme, setTheme };
}
