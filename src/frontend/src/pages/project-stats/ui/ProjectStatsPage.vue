<template>
  <div class="stats-page tab-fade-in">
    <!-- Filters & Actions Bar -->
    <div class="filters-card">
      <div class="filters-left">
        <div class="preset-group">
          <button
            type="button"
            class="preset-btn"
            :class="{ active: periodPreset === 'day' }"
            @click="selectPreset('day')"
          >
            {{ t('stats.day') }}
          </button>
          <button
            type="button"
            class="preset-btn"
            :class="{ active: periodPreset === 'week' }"
            @click="selectPreset('week')"
          >
            {{ t('stats.week') }}
          </button>
          <button
            type="button"
            class="preset-btn"
            :class="{ active: periodPreset === 'month' }"
            @click="selectPreset('month')"
          >
            {{ t('stats.month') }}
          </button>
          <button
            type="button"
            class="preset-btn"
            :class="{ active: periodPreset === 'quarter' }"
            @click="selectPreset('quarter')"
          >
            {{ t('stats.quarter') }}
          </button>
          <button
            type="button"
            class="preset-btn"
            :class="{ active: periodPreset === 'custom' }"
            @click="periodPreset = 'custom'"
          >
            {{ t('stats.custom') }}
          </button>
        </div>

        <div v-if="periodPreset === 'custom'" class="custom-range">
          <div class="date-field">
            <label>{{ t('stats.from') }}</label>
            <input type="date" v-model="customDateFrom" class="date-input" />
          </div>
          <div class="date-field">
            <label>{{ t('stats.to') }}</label>
            <input type="date" v-model="customDateTo" class="date-input" />
          </div>
          <button type="button" class="btn-primary-sm" @click="applyCustomDate">
            {{ t('stats.apply') }}
          </button>
        </div>
      </div>

      <!-- Right Actions (Refresh & Export) -->
      <div class="filter-actions">
        <button
          type="button"
          class="btn-secondary"
          :disabled="loading || refreshing"
          @click="loadData(true)"
        >
          <RotateCw class="icon-sm" :class="{ 'icon-spin': refreshing }" />
          <span>{{ t('stats.refresh') }}</span>
        </button>
        <button
          type="button"
          class="btn-secondary"
          :disabled="loading || !analyticsData"
          @click="exportCsv"
        >
          <Download class="icon-sm" />
          <span>{{ t('stats.exportCsv') }}</span>
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="state-container">
      <div class="spinner-wrap">
        <RotateCw class="icon-spin text-primary" style="width: 32px; height: 32px;" />
      </div>
      <p class="state-text">{{ t('stats.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-container">
      <AlertCircle class="error-icon" />
      <div class="error-content">
        <h3>{{ t('stats.errorTitle') }}</h3>
        <p>{{ error }}</p>
        <button type="button" class="btn-primary-sm" @click="loadData(false)">
          {{ t('stats.retry') }}
        </button>
      </div>
    </div>

    <!-- Main Content (Continuous Single-Page Scroll, No Tabs) -->
    <div v-else-if="analyticsData" class="stats-body">
      <!-- Top KPI Cards Grid (7 cards in single row on desktop) -->
      <div class="kpi-grid">
        <div class="kpi-card" :title="t('stats.kpi.players')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.players') }}</span>
            <Users class="kpi-icon text-blue" />
          </div>
          <div class="kpi-value">{{ formatNumber(summary?.unique_players) }}</div>
          <div class="kpi-sub">{{ t('stats.kpi.playersSub') }}</div>
        </div>

        <div class="kpi-card kpi-highlight" :title="t('stats.kpi.revenue')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.revenue') }}</span>
            <Wallet class="kpi-icon text-emerald" />
          </div>
          <div class="kpi-value text-emerald">₽ {{ formatNumber(summary?.total_revenue) }}</div>
          <div class="kpi-sub">
            {{ t('stats.kpi.revenueSub', { count: formatNumber(summary?.total_purchases) }) }}
          </div>
        </div>

        <div class="kpi-card" :title="t('stats.kpi.sessions')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.sessions') }}</span>
            <Gamepad2 class="kpi-icon text-purple" />
          </div>
          <div class="kpi-value">{{ formatNumber(summary?.total_sessions) }}</div>
          <div class="kpi-sub">
            {{ t('stats.kpi.sessionsSub', { spu: sessionsPerUser }) }}
          </div>
        </div>

        <div class="kpi-card" :title="t('stats.kpi.avgSession')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.avgSession') }}</span>
            <Clock class="kpi-icon text-amber" />
          </div>
          <div class="kpi-value">{{ (summary?.avg_session_minutes || 0).toFixed(1) }}m</div>
          <div class="kpi-sub">{{ t('stats.kpi.avgSessionSub') }}</div>
        </div>

        <div class="kpi-card" :title="t('stats.kpi.retentionD1')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.retentionD1') }}</span>
            <TrendingUp class="kpi-icon text-blue" />
          </div>
          <div class="kpi-value">{{ (summary?.d1_retention_rate || 0).toFixed(1) }}%</div>
          <div class="kpi-sub">{{ t('stats.kpi.retentionD1Sub') }}</div>
        </div>

        <div class="kpi-card" :title="t('stats.kpi.promoCtr')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.promoCtr') }}</span>
            <MousePointerClick class="kpi-icon text-cyan" />
          </div>
          <div class="kpi-value">{{ (summary?.overall_ctr || 0).toFixed(2) }}%</div>
          <div class="kpi-sub">
            {{
              t('stats.kpi.promoCtrSub', {
                clicks: formatNumber(summary?.total_promo_clicks),
                impressions: formatNumber(summary?.total_promo_impressions),
              })
            }}
          </div>
        </div>

        <div class="kpi-card" :title="t('stats.kpi.adImpressions')">
          <div class="kpi-header">
            <span class="kpi-label">{{ t('stats.kpi.adImpressions') }}</span>
            <Eye class="kpi-icon text-rose" />
          </div>
          <div class="kpi-value">{{ formatNumber(summary?.total_ad_impressions) }}</div>
          <div class="kpi-sub">{{ t('stats.kpi.adImpressionsSub') }}</div>
        </div>
      </div>

      <!-- SECTION 1: AUDIENCE & SESSIONS -->
      <section class="dashboard-section">
        <h2 class="dashboard-section-title">
          <Users class="section-icon text-blue" />
          <span>{{ t('stats.tabs.audience') }}</span>
        </h2>

        <div class="charts-2col">
          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.dauTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="dauChartData" :options="baseLineOptions" />
            </div>
          </div>

          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.sessionsTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="sessionsChartData" :options="dualAxisSessionsOptions" />
            </div>
          </div>
        </div>

        <div class="charts-2col">
          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.retentionTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="retentionChartData" :options="percentLineOptions" />
            </div>
          </div>

          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.churnTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="churnChartData" :options="percentLineOptions" />
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION 2: MONETIZATION -->
      <section class="dashboard-section">
        <h2 class="dashboard-section-title">
          <Wallet class="section-icon text-emerald" />
          <span>{{ t('stats.tabs.monetization') }}</span>
        </h2>

        <!-- Financial Metrics Mini Summary -->
        <div class="section-card">
          <h3 class="section-title">{{ t('stats.monetization.kpiTitle') }}</h3>
          <div class="mini-metrics-grid">
            <div class="mini-metric">
              <div class="mini-label">{{ t('stats.monetization.arpu') }}</div>
              <div class="mini-val">₽ {{ (summary?.arpu || 0).toFixed(1) }}</div>
              <div class="mini-sub">{{ t('stats.monetization.arpuSub') }}</div>
            </div>
            <div class="mini-metric">
              <div class="mini-label">{{ t('stats.monetization.arppu') }}</div>
              <div class="mini-val">₽ {{ (summary?.arppu || 0).toFixed(1) }}</div>
              <div class="mini-sub">{{ t('stats.monetization.arppuSub') }}</div>
            </div>
            <div class="mini-metric">
              <div class="mini-label">{{ t('stats.monetization.payingUsers') }}</div>
              <div class="mini-val">
                {{ formatNumber(summary?.paying_users_count) }}
                <span class="mini-sub-rate">({{ (summary?.paying_users_percent || 0).toFixed(1) }}%)</span>
              </div>
              <div class="mini-sub">{{ t('stats.monetization.avgCheck') }}: ₽ {{ (summary?.avg_order_value || 0).toFixed(0) }}</div>
            </div>
            <div class="mini-metric">
              <div class="mini-label">{{ t('stats.monetization.ltvSummary') }}</div>
              <div class="mini-val">₽ {{ (summary?.ltv || 0).toFixed(1) }}</div>
              <div class="mini-sub">{{ t('stats.monetization.ltvSummarySub') }}</div>
            </div>
            <div class="mini-metric">
              <div class="mini-label">{{ t('stats.monetization.timeToFirst') }}</div>
              <div class="mini-val">{{ (summary?.avg_hours_to_first_purchase || 0).toFixed(1) }} ч</div>
              <div class="mini-sub">{{ t('stats.monetization.timeToFirstSub') }}</div>
            </div>
          </div>
        </div>

        <div class="charts-2col">
          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.revenueTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Bar :data="(revenueWithPurchasesChartData as any)" :options="dualAxisRevenueOptions" />
            </div>
          </div>

          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.ltvTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="ltvChartData" :options="baseLineOptions" />
            </div>
          </div>
        </div>

        <!-- User Types & Repeat Purchases -->
        <div class="charts-2col">
          <!-- Structure: New vs Returning -->
          <div class="section-card">
            <h3 class="section-title">{{ t('stats.monetization.userTypesTitle') }}</h3>
            <div class="user-types-split">
              <div class="split-col">
                <div class="split-label">{{ t('stats.monetization.newUsers') }}</div>
                <div class="split-money">₽ {{ formatNumber(analyticsData.user_type_data?.new_users_revenue) }}</div>
                <div class="split-count">{{ formatNumber(analyticsData.user_type_data?.new_users_count) }} игроков</div>
              </div>
              <div class="split-col">
                <div class="split-label">{{ t('stats.monetization.returningUsers') }}</div>
                <div class="split-money">₽ {{ formatNumber(analyticsData.user_type_data?.returning_users_revenue) }}</div>
                <div class="split-count">{{ formatNumber(analyticsData.user_type_data?.returning_users_count) }} игроков</div>
              </div>
            </div>
            <!-- Split Bar -->
            <div class="split-bar-track">
              <div
                class="split-bar-fill new-users-fill"
                :style="{ width: `${newUsersRevenuePercent}%` }"
                :title="`Новые: ${newUsersRevenuePercent.toFixed(1)}%`"
              ></div>
              <div
                class="split-bar-fill returning-users-fill"
                :style="{ width: `${100 - newUsersRevenuePercent}%` }"
                :title="`Вернувшиеся: ${(100 - newUsersRevenuePercent).toFixed(1)}%`"
              ></div>
            </div>
            <div class="split-bar-legend">
              <div class="legend-item">
                <span class="legend-chip chip-cyan"></span>
                <span>{{ t('stats.monetization.newUsers') }}: {{ newUsersRevenuePercent.toFixed(1) }}%</span>
              </div>
              <div class="legend-item">
                <span class="legend-chip chip-emerald"></span>
                <span>{{ t('stats.monetization.returningUsers') }}: {{ (100 - newUsersRevenuePercent).toFixed(1) }}%</span>
              </div>
            </div>
          </div>

          <!-- RPR Segments -->
          <div class="section-card">
            <div class="section-header-flex">
              <h3 class="section-title">{{ t('stats.monetization.repeatPurchasesTitle') }}</h3>
              <span class="conversion-badge">
                {{ t('stats.monetization.repeatConversion', { rate: (summary?.rpr_conversion_rate1_to2 || 0).toFixed(1) }) }}
              </span>
            </div>
            <div class="rpr-list">
              <div v-for="seg in analyticsData.rpr_segments" :key="seg.segment_name" class="rpr-item">
                <div class="rpr-meta">
                  <span class="rpr-name">{{ seg.segment_name }}</span>
                  <span class="rpr-numbers">{{ formatNumber(seg.users_count) }} ({{ seg.percentage.toFixed(1) }}%)</span>
                </div>
                <div class="rpr-track">
                  <div class="rpr-fill" :style="{ width: `${seg.percentage}%` }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION 3: ADS & PROMO -->
      <section class="dashboard-section">
        <h2 class="dashboard-section-title">
          <Tv class="section-icon text-rose" />
          <span>{{ t('stats.tabs.adsPromo') }}</span>
        </h2>

        <div class="charts-2col">
          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.adsTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Bar :data="adsChartData" :options="stackedBarOptions" />
            </div>
          </div>

          <div class="chart-card">
            <div class="chart-header">
              <h3>{{ t('stats.charts.promoTitle') }}</h3>
            </div>
            <div class="chart-box">
              <Line :data="promoChartData" :options="percentLineOptions" />
            </div>
          </div>
        </div>

        <!-- Placements Performance Table -->
        <div class="section-card">
          <h3 class="section-title">{{ t('stats.promo.placementsTitle') }}</h3>
          <div class="table-responsive">
            <table class="analytics-table">
              <thead>
                <tr>
                  <th>{{ t('stats.promo.placement') }}</th>
                  <th class="text-right">{{ t('stats.promo.impressions') }}</th>
                  <th class="text-right">{{ t('stats.promo.clicks') }}</th>
                  <th class="text-right">{{ t('stats.promo.ctr') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="p in analyticsData.promo_placements" :key="p.placement">
                  <td class="font-medium">{{ p.placement_name || p.placement }}</td>
                  <td class="text-right">{{ formatNumber(p.impressions) }}</td>
                  <td class="text-right">{{ formatNumber(p.clicks) }}</td>
                  <td class="text-right">
                    <span class="ctr-badge" :class="getCtrBadgeClass(p.ctr)">
                      {{ p.ctr.toFixed(2) }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Users,
  Wallet,
  Clock,
  TrendingUp,
  MousePointerClick,
  Eye,
  Gamepad2,
  Tv,
  RotateCw,
  Download,
  AlertCircle,
} from 'lucide-vue-next';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  type ChartOptions,
  type ChartData,
} from 'chart.js';
import { Line, Bar } from 'vue-chartjs';

import { getProjectAnalytics } from '@/entities/analytics';
import type { GetProjectAnalyticsResponse } from '@/shared/types/generated/project_manager/v1/project';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const { t } = useI18n();
const route = useRoute();
const projectId = String(route.params.id || '');

// UI States
const loading = ref<boolean>(true);
const refreshing = ref<boolean>(false);
const error = ref<string>('');
const periodPreset = ref<'day' | 'week' | 'month' | 'quarter' | 'custom'>('week');

const customDateFrom = ref<string>('');
const customDateTo = ref<string>('');
const dateFrom = ref<string>('');
const dateTo = ref<string>('');

const analyticsData = ref<GetProjectAnalyticsResponse | null>(null);

const summary = computed(() => analyticsData.value?.summary);

const sessionsPerUser = computed(() => {
  if (!summary.value) return '0.0';
  const players = summary.value.unique_players || 1;
  return (summary.value.total_sessions / players).toFixed(1);
});

const newUsersRevenuePercent = computed(() => {
  if (!analyticsData.value?.user_type_data) return 50;
  const { new_users_revenue, returning_users_revenue } = analyticsData.value.user_type_data;
  const total = new_users_revenue + returning_users_revenue;
  if (total <= 0) return 50;
  return (new_users_revenue / total) * 100;
});

// Helpers
function formatNumber(num?: number | null): string {
  if (num === null || num === undefined) return '0';
  return new Intl.NumberFormat('ru-RU').format(Math.round(num));
}

function formatShortDate(dateStr?: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

function getCtrBadgeClass(ctr: number): string {
  if (ctr >= 6.0) return 'ctr-badge-success';
  if (ctr >= 3.5) return 'ctr-badge-primary';
  return 'ctr-badge-neutral';
}

function toISODate(d: Date): string {
  return d.toISOString().split('T')[0];
}

function selectPreset(preset: 'day' | 'week' | 'month' | 'quarter'): void {
  periodPreset.value = preset;
  const now = new Date();
  const past = new Date();
  if (preset === 'day') {
    past.setDate(now.getDate() - 1);
  } else if (preset === 'week') {
    past.setDate(now.getDate() - 7);
  } else if (preset === 'month') {
    past.setDate(now.getDate() - 30);
  } else if (preset === 'quarter') {
    past.setDate(now.getDate() - 90);
  }
  dateFrom.value = toISODate(past);
  dateTo.value = toISODate(now);
  loadData();
}

function applyCustomDate(): void {
  if (!customDateFrom.value || !customDateTo.value) return;
  dateFrom.value = customDateFrom.value;
  dateTo.value = customDateTo.value;
  loadData();
}

// Data loading
async function loadData(isRefresh = false): Promise<void> {
  if (!projectId) return;
  if (isRefresh) {
    refreshing.value = true;
  } else {
    loading.value = true;
  }
  error.value = '';

  try {
    const res = await getProjectAnalytics(projectId, {
      date_from: dateFrom.value,
      date_to: dateTo.value,
    });
    analyticsData.value = res;
  } catch (err: unknown) {
    console.error('Failed to load project analytics:', err);
    error.value = err instanceof Error ? err.message : 'Unknown error';
  } finally {
    loading.value = false;
    refreshing.value = false;
  }
}

// Chart.js Shared Options
const baseChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false,
  },
  plugins: {
    legend: {
      display: true,
      position: 'top' as const,
      labels: {
        color: '#8b949e',
        boxWidth: 10,
        boxHeight: 10,
        padding: 12,
        font: { family: 'inherit', size: 12 },
      },
    },
    tooltip: {
      backgroundColor: '#161b22',
      titleColor: '#f0f6fc',
      bodyColor: '#c9d1d9',
      borderColor: '#30363d',
      borderWidth: 1,
      padding: 10,
      cornerRadius: 8,
    },
  },
  scales: {
    x: {
      grid: {
        color: 'rgba(255, 255, 255, 0.05)',
      },
      ticks: {
        color: '#8b949e',
        font: { family: 'inherit', size: 11 },
      },
    },
    y: {
      grid: {
        color: 'rgba(255, 255, 255, 0.05)',
      },
      ticks: {
        color: '#8b949e',
        font: { family: 'inherit', size: 11 },
      },
      beginAtZero: true,
    },
  },
};

const baseLineOptions: ChartOptions<'line'> = {
  ...baseChartOptions,
  scales: {
    ...baseChartOptions.scales,
  },
};

const percentLineOptions: ChartOptions<'line'> = {
  ...baseChartOptions,
  scales: {
    ...baseChartOptions.scales,
    y: {
      ...baseChartOptions.scales.y,
      ticks: {
        color: '#8b949e',
        callback: (value) => `${value}%`,
        font: { family: 'inherit', size: 11 },
      },
    },
  },
};

const dualAxisSessionsOptions: ChartOptions<'line'> = {
  ...baseChartOptions,
  scales: {
    x: baseChartOptions.scales.x,
    y: {
      ...baseChartOptions.scales.y,
      position: 'left',
      title: {
        display: true,
        text: 'Сессии',
        color: '#8b949e',
      },
    },
    y1: {
      position: 'right',
      beginAtZero: true,
      grid: {
        drawOnChartArea: false,
      },
      ticks: {
        color: '#f59e0b',
        callback: (value) => `${value}м`,
      },
      title: {
        display: true,
        text: 'Ср. время',
        color: '#f59e0b',
      },
    },
  },
};

const dualAxisRevenueOptions: ChartOptions<'bar'> = {
  ...baseChartOptions,
  scales: {
    x: baseChartOptions.scales.x,
    y: {
      ...baseChartOptions.scales.y,
      position: 'left',
      ticks: {
        color: '#10b981',
        callback: (value) => `₽ ${value}`,
      },
    },
    y1: {
      position: 'right',
      beginAtZero: true,
      grid: {
        drawOnChartArea: false,
      },
      ticks: {
        color: '#38bdf8',
      },
      title: {
        display: true,
        text: 'Покупки',
        color: '#38bdf8',
      },
    },
  },
};

const stackedBarOptions: ChartOptions<'bar'> = {
  ...baseChartOptions,
  scales: {
    x: {
      ...baseChartOptions.scales.x,
      stacked: true,
    },
    y: {
      ...baseChartOptions.scales.y,
      stacked: true,
    },
  },
};

// Chart Data Computeds
const dauChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.dau_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.dauLegend'),
        data: items.map((i) => i.unique_players_count),
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.12)',
        fill: true,
        tension: 0.35,
        pointRadius: 3,
        pointHoverRadius: 6,
      },
    ],
  };
});

const revenueWithPurchasesChartData = computed<ChartData<'bar' | 'line'>>(() => {
  const items = analyticsData.value?.revenue_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        type: 'bar' as const,
        label: t('stats.charts.revenueLegend'),
        data: items.map((i) => i.total_revenue),
        backgroundColor: '#10b981',
        yAxisID: 'y',
        borderRadius: 4,
      },
      {
        type: 'line' as const,
        label: t('stats.charts.purchasesLegend'),
        data: items.map((i) => i.purchases_count),
        borderColor: '#38bdf8',
        backgroundColor: '#38bdf8',
        yAxisID: 'y1',
        pointRadius: 4,
        tension: 0.2,
      },
    ],
  };
});

const sessionsChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.session_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.sessionsCountLegend'),
        data: items.map((i) => i.total_sessions),
        borderColor: '#8b5cf6',
        backgroundColor: 'rgba(139, 92, 246, 0.15)',
        yAxisID: 'y',
        fill: true,
        tension: 0.3,
      },
      {
        label: t('stats.charts.sessionsDurationLegend'),
        data: items.map((i) => Number(i.avg_duration_minutes.toFixed(1))),
        borderColor: '#f59e0b',
        backgroundColor: '#f59e0b',
        yAxisID: 'y1',
        tension: 0.2,
        pointRadius: 4,
      },
    ],
  };
});

const retentionChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.retention_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: 'D1 Retention',
        data: items.map((i) => Number(i.d1.toFixed(1))),
        borderColor: '#3b82f6',
        tension: 0.2,
        pointRadius: 3,
      },
      {
        label: 'D3 Retention',
        data: items.map((i) => Number(i.d3.toFixed(1))),
        borderColor: '#10b981',
        tension: 0.2,
        pointRadius: 3,
      },
      {
        label: 'D7 Retention',
        data: items.map((i) => Number(i.d7.toFixed(1))),
        borderColor: '#f59e0b',
        tension: 0.2,
        pointRadius: 3,
      },
      {
        label: 'D30 Retention',
        data: items.map((i) => Number(i.d30.toFixed(1))),
        borderColor: '#ec4899',
        tension: 0.2,
        pointRadius: 3,
      },
    ],
  };
});

const churnChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.churn_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.churnLegend'),
        data: items.map((i) => Number(i.churn_rate.toFixed(1))),
        borderColor: '#f43f5e',
        backgroundColor: 'rgba(244, 63, 94, 0.1)',
        fill: true,
        tension: 0.3,
      },
    ],
  };
});

const ltvChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.ltv_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.ltvLegend'),
        data: items.map((i) => Number(i.ltv.toFixed(1))),
        borderColor: '#6366f1',
        backgroundColor: 'rgba(99, 102, 241, 0.15)',
        fill: true,
        tension: 0.3,
      },
    ],
  };
});

const adsChartData = computed<ChartData<'bar'>>(() => {
  const items = analyticsData.value?.ad_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.adInterstitial'),
        data: items.map((i) => i.interstitial),
        backgroundColor: '#f59e0b',
      },
      {
        label: t('stats.charts.adRewarded'),
        data: items.map((i) => i.rewarded),
        backgroundColor: '#10b981',
      },
      {
        label: t('stats.charts.adBanner'),
        data: items.map((i) => i.banner),
        backgroundColor: '#3b82f6',
      },
    ],
  };
});

const promoChartData = computed<ChartData<'line'>>(() => {
  const items = analyticsData.value?.promo_items || [];
  return {
    labels: items.map((i) => formatShortDate(i.date)),
    datasets: [
      {
        label: t('stats.charts.promoCtrLegend'),
        data: items.map((i) => Number(i.ctr.toFixed(2))),
        borderColor: '#06b6d4',
        backgroundColor: 'rgba(6, 182, 212, 0.1)',
        fill: true,
        tension: 0.3,
      },
    ],
  };
});

// CSV Export
function exportCsv(): void {
  if (!analyticsData.value) return;

  const data = analyticsData.value;
  const lines: string[] = [];

  lines.push('GDH Game Analytics Export');
  lines.push(`Project ID,${data.project_id}`);
  lines.push(`Date From,${data.date_from}`);
  lines.push(`Date To,${data.date_to}`);
  lines.push('');

  // Summary
  if (data.summary) {
    const s = data.summary;
    lines.push('--- SUMMARY ---');
    lines.push(`Unique Players,${s.unique_players}`);
    lines.push(`Total Revenue,${s.total_revenue}`);
    lines.push(`Total Sessions,${s.total_sessions}`);
    lines.push(`Avg Session Duration (min),${s.avg_session_minutes.toFixed(1)}`);
    lines.push(`D1 Retention Rate (%),${s.d1_retention_rate.toFixed(1)}`);
    lines.push(`Overall CTR (%),${s.overall_ctr.toFixed(2)}`);
    lines.push(`ARPU,${s.arpu.toFixed(2)}`);
    lines.push(`ARPPU,${s.arppu.toFixed(2)}`);
    lines.push(`Paying Users,${s.paying_users_count} (${s.paying_users_percent.toFixed(1)}%)`);
    lines.push(`Total Purchases,${s.total_purchases}`);
    lines.push(`Avg Order Value,${s.avg_order_value.toFixed(1)}`);
    lines.push(`LTV,${s.ltv.toFixed(1)}`);
    lines.push(`Total Ad Impressions,${s.total_ad_impressions}`);
    lines.push(`Total Promo Impressions,${s.total_promo_impressions}`);
    lines.push(`Total Promo Clicks,${s.total_promo_clicks}`);
    lines.push('');
  }

  // Daily Series
  lines.push('--- DAILY DATA ---');
  lines.push('Date,DAU,Revenue,Purchases,Sessions,AvgDurationMin,D1_Retention,D7_Retention,AdImpressions,PromoImpressions,PromoClicks,PromoCTR');

  const len = Math.max(
    data.dau_items.length,
    data.revenue_items.length,
    data.session_items.length,
    data.ad_items.length,
    data.promo_items.length
  );

  for (let i = 0; i < len; i++) {
    const dau = data.dau_items[i];
    const rev = data.revenue_items[i];
    const ses = data.session_items[i];
    const ret = data.retention_items[i];
    const ad = data.ad_items[i];
    const pro = data.promo_items[i];
    const dStr = dau?.date || rev?.date || `Day ${i + 1}`;

    lines.push([
      dStr,
      dau?.unique_players_count ?? 0,
      rev?.total_revenue ?? 0,
      rev?.purchases_count ?? 0,
      ses?.total_sessions ?? 0,
      ses?.avg_duration_minutes ? ses.avg_duration_minutes.toFixed(1) : '0.0',
      ret?.d1 ? ret.d1.toFixed(1) : '0.0',
      ret?.d7 ? ret.d7.toFixed(1) : '0.0',
      ad?.total ?? 0,
      pro?.impressions ?? 0,
      pro?.clicks ?? 0,
      pro?.ctr ? pro.ctr.toFixed(2) : '0.00',
    ].join(','));
  }
  lines.push('');

  // Promo Placements
  if (data.promo_placements.length > 0) {
    lines.push('--- PROMO PLACEMENTS ---');
    lines.push('Placement,Impressions,Clicks,CTR(%)');
    for (const p of data.promo_placements) {
      lines.push(`${p.placement_name || p.placement},${p.impressions},${p.clicks},${p.ctr.toFixed(2)}`);
    }
  }

  const csvContent = '\uFEFF' + lines.join('\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.setAttribute('href', url);
  link.setAttribute('download', `analytics-project-${projectId}-${dateFrom.value || 'start'}-${dateTo.value || 'end'}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// Lifecycle
onMounted(() => {
  selectPreset('week');
});

watch(() => route.params.id, (newId) => {
  if (newId && String(newId) !== projectId) {
    loadData();
  }
});
</script>

<style scoped>
.stats-page {
  padding: 16px 24px 48px 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.tab-fade-in {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Filters & Actions Bar */
.filters-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 10px;
  padding: 10px 14px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}

.filters-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.filter-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preset-group {
  display: flex;
  gap: 4px;
}

.preset-btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  background: transparent;
  color: var(--text-muted, #8b949e);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-btn:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-secondary, #21262d);
}

.preset-btn.active {
  color: #ffffff;
  background: var(--primary, #58a6ff);
}

.custom-range {
  display: flex;
  align-items: center;
  gap: 8px;
}

.date-field {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.84rem;
  color: var(--text-muted, #8b949e);
}

.date-input {
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  padding: 4px 8px;
  font-size: 0.84rem;
  font-family: inherit;
}

/* Top KPI Grid: 7 columns on desktop, 4 on medium screens, 2 on small screens */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 28px;
}

.kpi-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 10px;
  padding: 14px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  transition: border-color 0.2s ease, transform 0.2s ease;
}

.kpi-card:hover {
  border-color: rgba(88, 166, 255, 0.4);
}

.kpi-card.kpi-highlight {
  border-color: rgba(16, 185, 129, 0.4);
  background: linear-gradient(180deg, rgba(16, 185, 129, 0.06) 0%, var(--bg-card, #161b22) 100%);
}

.kpi-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 4px;
}

.kpi-label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-muted, #8b949e);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.kpi-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.kpi-value {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--text-main, #f0f6fc);
  letter-spacing: -0.3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.kpi-sub {
  font-size: 0.72rem;
  color: var(--text-muted, #8b949e);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Dashboard Sections */
.stats-body {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.dashboard-section {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.dashboard-section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border, #30363d);
}

.section-icon {
  width: 18px;
  height: 18px;
}

/* Charts & Grids */
.charts-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}

.chart-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
}

.chart-header {
  margin-bottom: 14px;
}

.chart-header h3 {
  margin: 0;
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.chart-box {
  width: 100%;
  height: 280px;
  position: relative;
}

/* Section Card */
.section-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 12px;
  padding: 18px;
}

.section-title {
  margin: 0 0 14px 0;
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.section-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.section-header-flex .section-title {
  margin: 0;
}

.conversion-badge {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 0.78rem;
  font-weight: 600;
}

/* Mini Metrics Grid */
.mini-metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}

.mini-metric {
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  padding: 12px;
}

.mini-label {
  font-size: 0.78rem;
  color: var(--text-muted, #8b949e);
  margin-bottom: 4px;
}

.mini-val {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  margin-bottom: 2px;
}

.mini-sub-rate {
  font-size: 0.8rem;
  color: #10b981;
  font-weight: 500;
}

.mini-sub {
  font-size: 0.7rem;
  color: var(--text-muted, #8b949e);
}

/* User Types Split */
.user-types-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
}

.split-col {
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  padding: 12px;
}

.split-label {
  font-size: 0.78rem;
  color: var(--text-muted, #8b949e);
  margin-bottom: 4px;
}

.split-money {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.split-count {
  font-size: 0.74rem;
  color: var(--text-muted, #8b949e);
}

.split-bar-track {
  width: 100%;
  height: 8px;
  background: var(--bg-secondary, #21262d);
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  margin-bottom: 10px;
}

.split-bar-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.new-users-fill {
  background: #06b6d4;
}

.returning-users-fill {
  background: #10b981;
}

.split-bar-legend {
  display: flex;
  gap: 16px;
  font-size: 0.78rem;
  color: var(--text-muted, #8b949e);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-chip {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

.chip-cyan {
  background: #06b6d4;
}

.chip-emerald {
  background: #10b981;
}

/* RPR List */
.rpr-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.rpr-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.rpr-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.82rem;
}

.rpr-name {
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
}

.rpr-numbers {
  color: var(--text-muted, #8b949e);
}

.rpr-track {
  width: 100%;
  height: 6px;
  background: var(--bg-secondary, #21262d);
  border-radius: 3px;
  overflow: hidden;
}

.rpr-fill {
  height: 100%;
  background: #8b5cf6;
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* Table */
.table-responsive {
  overflow-x: auto;
}

.analytics-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.88rem;
}

.analytics-table th {
  text-align: left;
  padding: 10px 12px;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
  border-bottom: 1px solid var(--border, #30363d);
}

.analytics-table td {
  padding: 10px 12px;
  color: var(--text-main, #f0f6fc);
  border-bottom: 1px solid var(--border, #30363d);
}

.analytics-table tbody tr:last-child td {
  border-bottom: none;
}

.text-right {
  text-align: right !important;
}

.font-medium {
  font-weight: 500;
}

.ctr-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.78rem;
  font-weight: 600;
}

.ctr-badge-success {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.ctr-badge-primary {
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
}

.ctr-badge-neutral {
  background: rgba(139, 148, 158, 0.15);
  color: #8b949e;
}

/* Buttons */
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  border-radius: 6px;
  color: var(--text-main, #f0f6fc);
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-tertiary, #30363d);
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary-sm {
  padding: 5px 12px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: 1px solid var(--primary, #58a6ff);
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.btn-primary-sm:hover {
  background: var(--primary-hover, #4a94ec);
}

/* States */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 20px;
  gap: 16px;
}

.state-text {
  color: var(--text-muted, #8b949e);
  font-size: 0.95rem;
}

.error-container {
  background: rgba(248, 81, 73, 0.08);
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: 10px;
  padding: 20px;
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.error-icon {
  width: 22px;
  height: 22px;
  color: #f85149;
  flex-shrink: 0;
}

.error-content h3 {
  margin: 0 0 6px 0;
  color: #f85149;
  font-size: 0.98rem;
}

.error-content p {
  margin: 0 0 12px 0;
  color: var(--text-main, #f0f6fc);
  font-size: 0.88rem;
}

/* Color classes & animations */
.icon-sm {
  width: 15px;
  height: 15px;
}

.icon-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.text-blue {
  color: #3b82f6;
}

.text-emerald {
  color: #10b981;
}

.text-purple {
  color: #8b5cf6;
}

.text-amber {
  color: #f59e0b;
}

.text-cyan {
  color: #06b6d4;
}

.text-rose {
  color: #f43f5e;
}

/* Responsive Breakpoints */
@media (max-width: 1200px) {
  .kpi-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 900px) {
  .charts-2col {
    grid-template-columns: 1fr;
  }
  .filters-card {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-actions {
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>
