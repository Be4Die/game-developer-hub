export type ModerationRequestStatus =
  | 'REQUEST_STATUS_UNSPECIFIED'
  | 'REQUEST_STATUS_PENDING'
  | 'REQUEST_STATUS_IN_REVIEW'
  | 'REQUEST_STATUS_APPROVED'
  | 'REQUEST_STATUS_REJECTED'
  | 'REQUEST_STATUS_CANCELLED'
  | number
  | string;

export type ModerationRequestType =
  | 'REQUEST_TYPE_UNSPECIFIED'
  | 'REQUEST_TYPE_PROJECT_PUBLICATION'
  | 'REQUEST_TYPE_SERVER_ACCESS'
  | 'REQUEST_TYPE_PROJECT_UPDATE'
  | number
  | string;

export interface GameItemSnapshot {
  game_item_id: string;
  name: string;
  description?: string;
  price_coins: number;
  image_url?: string;
  is_active?: boolean;
}

export interface ModerationViolation {
  rule_id?: string;
  rule_title?: string;
  rule_number?: string;
  comment?: string;
  [key: string]: any;
}

export interface ModerationRequest {
  id: string | number;
  projectId?: string | number;
  project_id?: string | number;
  type: ModerationRequestType;
  status: ModerationRequestStatus;
  ownerId?: string;
  owner_id?: string;
  moderatorId?: string;
  moderator_id?: string;
  submittedAt?: string;
  submitted_at?: string;
  resolvedAt?: string;
  resolved_at?: string;
  reason?: string;
  moderatorComment?: string;
  moderator_comment?: string;
  rejectionReason?: string;
  rejection_reason?: string;
  violations?: ModerationViolation[];
  snapshot?: Record<string, any>;
  maxInstances?: number;
  max_instances?: number;
  maxTotalCpuMillis?: number;
  max_total_cpu_millis?: number;
  maxTotalMemoryMb?: number;
  max_total_memory_mb?: number;
  maxInstanceCpuMillis?: number;
  max_instance_cpu_millis?: number;
  maxInstanceMemoryMb?: number;
  max_instance_memory_mb?: number;
  [key: string]: any;
}

export interface ChatAttachment {
  id: string;
  file_name: string;
  file_size?: number;
  mime_type?: string;
  url: string;
  thumbnail_url?: string;
  created_at?: string;
  [key: string]: any;
}

export interface ChatMessage {
  id: string;
  project_id?: string | number;
  sender_id?: string;
  sender_name?: string;
  sender_role?: string;
  content: string;
  attachment_ids?: string[];
  attachments?: ChatAttachment[];
  payload_json?: string;
  created_at?: string;
  [key: string]: any;
}

export interface ActiveChat {
  project_id: string | number;
  project_title?: string;
  unread_count?: number;
  last_message?: ChatMessage;
  status?: string;
  updated_at?: string;
  [key: string]: any;
}

export interface ModeratorStats {
  moderator_id?: string;
  display_name?: string;
  approved_count?: number;
  rejected_count?: number;
  total_reviewed?: number;
  avg_review_time_seconds?: number;
  [key: string]: any;
}
