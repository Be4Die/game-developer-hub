export interface ClientBuild {
  id?: number | string;
  project_id?: number | string;
  version: string;
  size_bytes?: number;
  entrypoint?: string;
  created_at?: string;
  [key: string]: any;
}

export interface ServerBuild {
  id?: string;
  game_id?: string | number;
  version?: string;
  build_version?: string;
  image?: string;
  image_tag?: string;
  protocol?: string;
  internal_port?: number;
  max_players?: number;
  file_size_bytes?: number;
  size_bytes?: number;
  created_at?: string;
  [key: string]: any;
}

export type Build = ServerBuild;
