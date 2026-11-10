| Сервис | Число операций | Используемые интерфейсы и сервисы |
|---|---|---|
| **Оркестрация** | | |
| `NodeService` | 26 | `NodeRepo`, `NodeStateStore`, `InstanceRepo`, `InstanceStateStore`, `NodeClient`, `ManagedServiceRepo`, `BackupRepo`, `PlatformAccessRepo` |
| `InstanceService` | 9 | `InstanceRepo`, `InstanceStateStore`, `BuildStorage`, `NodeRepo`, `NodeStateStore`, `NodeClient`, `ManagedServiceRepo`, `PlatformAccessRepo` |
| `BuildPipeline` | 4 | `BuildStorage`, `BuildStorageFS`, `NodeClient`, `NodeRepo`, `NodeStateStore` |
| `GamePolicyService` | 2 | `GamePolicyRepo` |
| `DiscoveryService` | 1 | `InstanceRepo`, `InstanceStateStore`, `NodeRepo`, `BuildStorage`, `GamePolicyService` |
| `QueueService` | 7 | `QueueStore`, `QueueEventRepo`, `GamePolicyService`, `InstanceRepo`, `InstanceStateStore`, `NodeRepo` |
| `HeartbeatService` | 3 | `NodeRepo`, `NodeStateStore`, `InstanceRepo`, `InstanceStateStore`, `NodeClient`, `BuildStorage`, `GamePolicyService`, `QueueService` |
| `PlatformAccessService` | 3 | `PlatformAccessRepo` |
| **Управление проектами** | | |
| `ProjectService` | 39 | `ProjectRepo`, `DraftRepo`, `BuildRepo`, `ReleaseRepo`, `DeploymentRepo`, `MemberRepo`, `InvitationRepo`, `BlockRepo`, `ModerationClient`, `PurchaseClient`, `AnalyticsClient`, `BuildStorage`, `MediaStorage`, `Deployer`, `Locker` |
| **Модерация** | | |
| `ModerationService` | 21 | `RequestRepo`, `MessageRepo`, `AttachmentRepo`, `SnapshotRepo`, `ProjectClient`, `OrchestratorClient` |
| **Учётные записи** | | |
| `AuthService` | 8 | `UserRepository`, `SessionRepository`, `TokenManager`, `PasswordHasher`, `EmailVerificationStore`, `PasswordResetStore`, `EmailSender`, `AttemptLimiter` |
| `UserService` | 8 | `UserRepository`, `SessionRepository`, `PasswordHasher` |
| `TokenService` | 4 | `SessionRepository`, `SessionCache`, `TokenManager` |
| **Агент узла** | | |
| `DeploymentService` | 26 | `InstanceStorage`, `ContainerRuntime` |
| `DiscoveryService` | 6 | `InstanceStorage`, `ContainerRuntime` |
| `BackupManager` | 7 | `ContainerRuntime` |
| `AnnouncementService` | 2 | – |
