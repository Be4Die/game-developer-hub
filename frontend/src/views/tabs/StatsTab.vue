<template>
  <div class="tab-content tab-fade-in">
    <h1 class="page-title">Статистика</h1>

    <!-- Карточки метрик -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-header">
          <span class="label">Игроков за сутки</span>
          <Users class="icon-sm stat-icon" />
        </div>
        <span class="value">9,341</span>
        <span class="trend up">+12.5% за неделю</span>
      </div>
      <div class="stat-card">
        <div class="stat-header">
          <span class="label">Среднее время</span>
          <Clock class="icon-sm stat-icon" />
        </div>
        <span class="value">13m</span>
        <span class="trend down">-2.1% за неделю</span>
      </div>
      <div class="stat-card">
        <div class="stat-header">
          <span class="label">Примерный доход</span>
          <Wallet class="icon-sm stat-icon" />
        </div>
        <span class="value">₽ 34 500</span>
        <span class="trend up">+8.3% за неделю</span>
      </div>
    </div>

    <!-- График: Активность игроков -->
    <div class="chart-card">
      <div class="chart-header">
        <h3>Активность игроков (7 дней)</h3>
        <div class="chart-legend">
          <span class="legend-dot blue"></span> Игроки
        </div>
      </div>
      <div class="chart-container">
        <svg viewBox="0 0 700 250" class="line-chart">
          <!-- Сетка -->
          <line x1="50" y1="20" x2="50" y2="220" stroke="var(--border)" stroke-width="1" />
          <line x1="50" y1="220" x2="680" y2="220" stroke="var(--border)" stroke-width="1" />
          <line x1="50" y1="170" x2="680" y2="170" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="120" x2="680" y2="120" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="70" x2="680" y2="70" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="20" x2="680" y2="20" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />

          <!-- Y-axis labels -->
          <text x="45" y="224" text-anchor="end" class="axis-label">0</text>
          <text x="45" y="174" text-anchor="end" class="axis-label">2.5k</text>
          <text x="45" y="124" text-anchor="end" class="axis-label">5k</text>
          <text x="45" y="74" text-anchor="end" class="axis-label">7.5k</text>
          <text x="45" y="24" text-anchor="end" class="axis-label">10k</text>

          <!-- Область под графиком -->
          <defs>
            <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3B82F6" stop-opacity="0.3" />
              <stop offset="100%" stop-color="#3B82F6" stop-opacity="0.02" />
            </linearGradient>
          </defs>
          <path d="M140,95 L235,60 L330,80 L425,40 L520,55 L615,30 L615,220 L140,220 Z" fill="url(#areaGradient)" />

          <!-- Линия -->
          <polyline points="140,95 235,60 330,80 425,40 520,55 615,30" fill="none" stroke="#3B82F6" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />

          <!-- Точки -->
          <circle cx="140" cy="95" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />
          <circle cx="235" cy="60" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />
          <circle cx="330" cy="80" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />
          <circle cx="425" cy="40" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />
          <circle cx="520" cy="55" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />
          <circle cx="615" cy="30" r="5" fill="#3B82F6" stroke="white" stroke-width="2" />

          <!-- X-axis labels -->
          <text x="140" y="240" text-anchor="middle" class="axis-label">Пн</text>
          <text x="235" y="240" text-anchor="middle" class="axis-label">Вт</text>
          <text x="330" y="240" text-anchor="middle" class="axis-label">Ср</text>
          <text x="425" y="240" text-anchor="middle" class="axis-label">Чт</text>
          <text x="520" y="240" text-anchor="middle" class="axis-label">Пт</text>
          <text x="615" y="240" text-anchor="middle" class="axis-label">Сб</text>
        </svg>
      </div>
    </div>

    <!-- График: Доход -->
    <div class="chart-card">
      <div class="chart-header">
        <h3>Доход за неделю (₽)</h3>
        <div class="chart-legend">
          <span class="legend-dot green"></span> Доход
        </div>
      </div>
      <div class="chart-container">
        <svg viewBox="0 0 700 250" class="bar-chart">
          <!-- Сетка -->
          <line x1="50" y1="20" x2="50" y2="220" stroke="var(--border)" stroke-width="1" />
          <line x1="50" y1="220" x2="680" y2="220" stroke="var(--border)" stroke-width="1" />
          <line x1="50" y1="170" x2="680" y2="170" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="120" x2="680" y2="120" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="70" x2="680" y2="70" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />
          <line x1="50" y1="20" x2="680" y2="20" stroke="var(--border)" stroke-width="0.5" stroke-dasharray="4,4" />

          <!-- Y-axis labels -->
          <text x="45" y="224" text-anchor="end" class="axis-label">0</text>
          <text x="45" y="174" text-anchor="end" class="axis-label">5k</text>
          <text x="45" y="124" text-anchor="end" class="axis-label">10k</text>
          <text x="45" y="74" text-anchor="end" class="axis-label">15k</text>
          <text x="45" y="24" text-anchor="end" class="axis-label">20k</text>

          <!-- Столбцы -->
          <rect x="100" y="140" width="50" height="80" rx="4" fill="#10B981" opacity="0.8" />
          <text x="125" y="132" text-anchor="middle" class="bar-value">4,200</text>

          <rect x="195" y="110" width="50" height="110" rx="4" fill="#10B981" opacity="0.8" />
          <text x="220" y="102" text-anchor="middle" class="bar-value">5,800</text>

          <rect x="290" y="90" width="50" height="130" rx="4" fill="#10B981" opacity="0.8" />
          <text x="315" y="82" text-anchor="middle" class="bar-value">6,500</text>

          <rect x="385" y="60" width="50" height="160" rx="4" fill="#10B981" opacity="0.8" />
          <text x="410" y="52" text-anchor="middle" class="bar-value">7,900</text>

          <rect x="480" y="80" width="50" height="140" rx="4" fill="#10B981" opacity="0.8" />
          <text x="505" y="72" text-anchor="middle" class="bar-value">7,100</text>

          <rect x="575" y="35" width="50" height="185" rx="4" fill="#10B981" opacity="0.8" />
          <text x="600" y="27" text-anchor="middle" class="bar-value">9,000</text>

          <!-- X-axis labels -->
          <text x="125" y="240" text-anchor="middle" class="axis-label">Пн</text>
          <text x="220" y="240" text-anchor="middle" class="axis-label">Вт</text>
          <text x="315" y="240" text-anchor="middle" class="axis-label">Ср</text>
          <text x="410" y="240" text-anchor="middle" class="axis-label">Чт</text>
          <text x="505" y="240" text-anchor="middle" class="axis-label">Пт</text>
          <text x="600" y="240" text-anchor="middle" class="axis-label">Сб</text>
        </svg>
      </div>
    </div>

    <!-- Кольцевая диаграмма: Источники трафика -->
    <div class="charts-row">
      <div class="chart-card donut-card">
        <div class="chart-header">
          <h3>Источники трафика</h3>
        </div>
        <div class="donut-container">
          <svg viewBox="0 0 200 200" class="donut-chart">
            <circle cx="100" cy="100" r="70" fill="none" stroke="#E5E7EB" stroke-width="30" />
            <!-- Поиск: 45% -->
            <circle cx="100" cy="100" r="70" fill="none" stroke="#3B82F6" stroke-width="30"
              stroke-dasharray="198 242" stroke-dashoffset="0"
              transform="rotate(-90 100 100)" />
            <!-- Прямой: 30% -->
            <circle cx="100" cy="100" r="70" fill="none" stroke="#10B981" stroke-width="30"
              stroke-dasharray="132 308" stroke-dashoffset="-198"
              transform="rotate(-90 100 100)" />
            <!-- Соцсети: 25% -->
            <circle cx="100" cy="100" r="70" fill="none" stroke="#F59E0B" stroke-width="30"
              stroke-dasharray="110 330" stroke-dashoffset="-330"
              transform="rotate(-90 100 100)" />
            <!-- Центр -->
            <circle cx="100" cy="100" r="50" fill="var(--bg-card)" />
            <text x="100" y="95" text-anchor="middle" class="donut-value">9,341</text>
            <text x="100" y="115" text-anchor="middle" class="donut-sub">всего</text>
          </svg>
        </div>
        <div class="donut-legend">
          <div class="legend-item">
            <span class="legend-dot blue"></span>
            <span class="legend-label">Поиск</span>
            <span class="legend-percent">45%</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot green"></span>
            <span class="legend-label">Прямой</span>
            <span class="legend-percent">30%</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot yellow"></span>
            <span class="legend-label">Соцсети</span>
            <span class="legend-percent">25%</span>
          </div>
        </div>
      </div>

      <!-- Мини-таблица: Топ платформы -->
      <div class="chart-card table-card">
        <div class="chart-header">
          <h3>Топ платформы</h3>
        </div>
        <div class="platform-table">
          <div class="table-row header">
            <span>Платформа</span>
            <span>Игроки</span>
            <span>Доля</span>
          </div>
          <div class="table-row">
            <div class="platform-name"><Monitor class="icon-xs" /> PC (Windows)</div>
            <span>5,604</span>
            <span><span class="bar-mini" style="width: 60%"></span>60%</span>
          </div>
          <div class="table-row">
            <div class="platform-name"><Smartphone class="icon-xs" /> Android</div>
            <span>2,802</span>
            <span><span class="bar-mini" style="width: 30%"></span>30%</span>
          </div>
          <div class="table-row">
            <div class="platform-name"><Tablet class="icon-xs" /> iOS</div>
            <span>935</span>
            <span><span class="bar-mini" style="width: 10%"></span>10%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Users, Clock, Wallet, Monitor, Smartphone, Tablet } from 'lucide-vue-next'
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.tab-content { padding: 32px; max-width: 1000px; }
.page-title { margin: 0 0 24px 0; font-size: 1.5rem; }

/* Карточки метрик */
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 24px; }
.stat-card { background: var(--bg-card); padding: 24px; border-radius: 12px; border: 1px solid var(--border); display: flex; flex-direction: column; gap: 8px; }
.stat-header { display: flex; justify-content: space-between; align-items: center; }
.stat-icon { color: var(--text-muted); }
.label { color: var(--text-muted); font-weight: 500; font-size: 0.9rem; }
.value { font-size: 2rem; font-weight: 800; color: var(--text-main); }
.trend { font-size: 0.8rem; font-weight: 600; }
.trend.up { color: var(--success); }
.trend.down { color: var(--danger); }

/* Общие стили графиков */
.chart-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 24px; margin-bottom: 24px; }
.chart-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.chart-header h3 { margin: 0; font-size: 1rem; font-weight: 600; }
.chart-legend { display: flex; align-items: center; gap: 6px; font-size: 0.85rem; color: var(--text-muted); }
.legend-dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; }
.legend-dot.blue { background: #3B82F6; }
.legend-dot.green { background: #10B981; }
.legend-dot.yellow { background: #F59E0B; }

.chart-container { width: 100%; }
.axis-label { fill: var(--text-muted); font-size: 11px; font-family: inherit; }

/* Линейный график */
.line-chart { width: 100%; height: auto; }
.bar-value { fill: var(--text-muted); font-size: 10px; font-family: inherit; }

/* Кольцевая диаграмма */
.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.donut-card { margin-bottom: 0; }
.donut-container { display: flex; justify-content: center; margin-bottom: 20px; }
.donut-chart { width: 200px; height: 200px; transform: rotate(0deg); }
.donut-value { fill: var(--text-main); font-size: 20px; font-weight: 800; font-family: inherit; }
.donut-sub { fill: var(--text-muted); font-size: 12px; font-family: inherit; }
.donut-legend { display: flex; flex-direction: column; gap: 8px; }
.legend-item { display: flex; align-items: center; gap: 8px; font-size: 0.85rem; }
.legend-label { flex: 1; color: var(--text-main); }
.legend-percent { font-weight: 600; color: var(--text-muted); }

/* Таблица платформ */
.table-card { margin-bottom: 0; }
.platform-table { display: flex; flex-direction: column; }
.table-row { display: flex; align-items: center; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--border); font-size: 0.9rem; color: var(--text-main); }
.table-row:last-child { border-bottom: none; }
.table-row.header { color: var(--text-muted); font-weight: 600; font-size: 0.8rem; padding-bottom: 8px; }
.platform-name { display: flex; align-items: center; gap: 8px; }
.icon-xs { width: 14px; height: 14px; color: var(--text-muted); }
.bar-mini { display: inline-block; height: 6px; background: var(--primary); border-radius: 3px; margin-right: 6px; vertical-align: middle; }

/* Адаптивность */
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr; }
  .charts-row { grid-template-columns: 1fr; }
}
</style>
