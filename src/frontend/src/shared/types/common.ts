export interface PaginationParams {
  limit?: number;
  offset?: number;
  page?: number;
  pageSize?: number;
  query?: string;
}

export interface PaginatedResponse<T> {
  items?: T[];
  data?: T[];
  total_count?: number;
  total?: number;
  page?: number;
  pageSize?: number;
}

export interface ApiErrorResponse {
  code?: number | string;
  message?: string;
  error?: string;
  details?: unknown[];
}
