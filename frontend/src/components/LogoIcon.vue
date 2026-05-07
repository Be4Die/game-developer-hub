<template>
    <div class="logo-wrapper" :class="{ 'no-hover': noHover }">
        <div class="icon-bg" :style="{ width: size + 'px', height: size + 'px' }">
            <svg
                class="hub-icon"
                viewBox="0 0 64 64"
                xmlns="http://www.w3.org/2000/svg"
            >
                <!-- 6 outer nodes -->
                <circle class="node" cx="32" cy="8" r="4" />
                <circle class="node" cx="52.8" cy="20" r="4" />
                <circle class="node" cx="52.8" cy="44" r="4" />
                <circle class="node" cx="32" cy="56" r="4" />
                <circle class="node" cx="11.2" cy="44" r="4" />
                <circle class="node" cx="11.2" cy="20" r="4" />

                <!-- center glow ring -->
                <circle class="center-glow" cx="32" cy="32" r="10" />
                <!-- center core -->
                <circle class="center-core" cx="32" cy="32" r="5.5" />

                <!-- connections -->
                <line class="conn" x1="32" y1="8" x2="32" y2="32" />
                <line class="conn" x1="52.8" y1="20" x2="32" y2="32" />
                <line class="conn" x1="52.8" y1="44" x2="32" y2="32" />
                <line class="conn" x1="32" y1="56" x2="32" y2="32" />
                <line class="conn" x1="11.2" y1="44" x2="32" y2="32" />
                <line class="conn" x1="11.2" y1="20" x2="32" y2="32" />
            </svg>
        </div>

        <div v-if="showText" class="text-container">
            <div
                class="logo-main"
                :style="{ fontSize: textSize + 'px', letterSpacing: textSize > 20 ? '3px' : '2px' }"
            >
                WELWISE
            </div>
            <div v-if="showSub" class="logo-sub">Developer Platform</div>
        </div>
    </div>
</template>

<script setup>
defineProps({
    size: { type: Number, default: 32 },
    textSize: { type: Number, default: 18 },
    showText: { type: Boolean, default: true },
    showSub: { type: Boolean, default: false },
    noHover: { type: Boolean, default: false },
});
</script>

<style scoped>
.logo-wrapper {
    display: flex;
    align-items: center;
    gap: 10px;
}

.icon-bg {
    background: linear-gradient(135deg, #3b5bdb 0%, #5c7cfa 100%);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(59, 91, 219, 0.35);
    transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.logo-wrapper:not(.no-hover):hover .icon-bg {
    transform: scale(1.06) rotate(3deg);
    box-shadow: 0 4px 16px rgba(59, 91, 219, 0.5);
}

.hub-icon {
    width: 68%;
    height: 68%;
    display: block;
}

.node {
    fill: white;
    opacity: 0.9;
}

.center-glow {
    fill: white;
    opacity: 0.15;
}

.center-core {
    fill: white;
}

.conn {
    stroke: white;
    stroke-width: 2.2;
    stroke-linecap: round;
    opacity: 0.55;
}

.text-container {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.logo-main {
    font-weight: 800;
    letter-spacing: 2px;
    color: var(--text-main);
    line-height: 1.1;
    transition: color 0.2s;
}

.logo-sub {
    font-size: 10px;
    font-weight: 500;
    letter-spacing: 2px;
    color: var(--text-muted);
    text-transform: uppercase;
}

.logo-wrapper:not(.no-hover):hover .logo-main {
    color: var(--primary);
}
</style>
