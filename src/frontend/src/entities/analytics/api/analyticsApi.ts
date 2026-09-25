import { http } from '@/shared/api';
import type { GetProjectAnalyticsResponse } from '@/shared/types/generated/project_manager/v1/project';

export interface AnalyticsFilterParams {
  date_from?: string;
  date_to?: string;
  aggregation_period?: 'Day' | 'Week' | 'Month';
}

export const getProjectAnalytics = (
  projectId: string | number,
  params: AnalyticsFilterParams = {}
): Promise<GetProjectAnalyticsResponse> => {
  const query = new URLSearchParams();
  if (params.date_from) query.append('date_from', params.date_from);
  if (params.date_to) query.append('date_to', params.date_to);
  if (params.aggregation_period) query.append('aggregation_period', params.aggregation_period);

  const qStr = query.toString() ? `?${query.toString()}` : '';
  return http.get(`/projects/${projectId}/analytics${qStr}`).then((r) => r.data);
};
