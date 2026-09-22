-- 06_project_update_type.sql — тип заявки для обновления опубликованного проекта
COMMENT ON COLUMN moderation_requests.request_type IS '1 = REQUEST_TYPE_PROJECT_PUBLICATION, 2 = REQUEST_TYPE_SERVER_ACCESS, 3 = REQUEST_TYPE_PROJECT_UPDATE';
