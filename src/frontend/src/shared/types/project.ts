export type ProjectStatus =
  | 'PROJECT_STATUS_UNSPECIFIED'
  | 'PROJECT_STATUS_DRAFT'
  | 'PROJECT_STATUS_PENDING'
  | 'PROJECT_STATUS_PUBLISHED'
  | 'PROJECT_STATUS_REJECTED'
  | number;

export interface ProjectDraft {
  project_id?: number | string;
  title_ru?: string;
  title_en?: string;
  seo_ru?: string;
  seo_en?: string;
  about_ru?: string;
  about_en?: string;
  icon_path?: string;
  cover_path?: string;
  video_path?: string;
  active_build_version?: string;
  dev_url?: string;
  updated_at?: string;
  is_online?: boolean;
}

export interface ProjectRelease {
  id?: number | string;
  project_id?: number | string;
  version?: string;
  title_ru?: string;
  title_en?: string;
  seo_ru?: string;
  seo_en?: string;
  about_ru?: string;
  about_en?: string;
  icon_path?: string;
  cover_path?: string;
  video_path?: string;
  prod_url?: string;
  published_at?: string;
  is_online?: boolean;
}

export interface Project {
  id: number | string;
  owner_id?: string;
  title_ru?: string;
  title_en?: string;
  seo_ru?: string;
  seo_en?: string;
  about_ru?: string;
  about_en?: string;
  status: ProjectStatus;
  icon_path?: string;
  cover_path?: string;
  video_path?: string;
  active_build_version?: string;
  dev_url?: string;
  prod_url?: string;
  created_at?: string;
  updated_at?: string;
  draft?: ProjectDraft;
  release?: ProjectRelease;
  current_user_permissions?: string[];
  is_owner?: boolean;
  is_online?: boolean;
  [key: string]: any;
}

export interface ProjectInvitation {
  id: string | number;
  project_id: string | number;
  inviter_id?: string;
  invitee_id?: string;
  invitee_email?: string;
  status?: string;
  permissions?: string[];
  created_at?: string;
}

export interface ProjectMember {
  user_id: string;
  email?: string;
  display_name?: string;
  permissions?: string[];
  joined_at?: string;
}

export interface BlockedUser {
  id?: string;
  blocked_user_id: string;
  blocked_email?: string;
  created_at?: string;
}
