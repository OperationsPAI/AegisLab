// ==================== Common Types ====================

export interface Label {
  key: string
  value: string
}

export interface PaginationParams {
  page: number
  size: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  size: number
}

// ==================== Project State Enums ====================

export enum ProjectState {
  ACTIVE = 0,
  PAUSED = 1,
  COMPLETED = 2,
  ARCHIVED = 3,
}

// ==================== Task State Enums ====================

export enum TaskState {
  PENDING = 0,
  RUNNING = 1,
  COMPLETED = 2,
  ERROR = 3,
  CANCELLED = 4,
}

export enum InjectionState {
  PENDING = 0,
  RUNNING = 1,
  COMPLETED = 2,
  ERROR = 3,
  STOPPED = 4,
}

export enum InjectionType {
  NETWORK = 0,
  CPU = 1,
  MEMORY = 2,
  DISK = 3,
  PROCESS = 4,
  KUBERNETES = 5,
}

export enum ExecutionState {
  PENDING = 0,
  RUNNING = 1,
  COMPLETED = 2,
  ERROR = -1,
}

// ==================== Container Types ====================

export enum ContainerType {
  PEDESTAL = 'Pedestal',
  BENCHMARK = 'Benchmark',
  ALGORITHM = 'Algorithm',
}

export interface Container {
  id: number
  name: string
  type: ContainerType
  readme: string
  is_public: boolean
  created_at: string
  updated_at: string
  labels?: Label[]
  versions?: ContainerVersion[]
}

export interface ContainerVersion {
  id: number
  container_id: number
  version: string
  registry: string
  repository: string
  tag: string
  command?: string
  created_at: string
}

// ==================== Dataset Types ====================

export enum DatasetType {
  TRACE = 'Trace',
  LOG = 'Log',
  METRIC = 'Metric',
}

export interface Dataset {
  id: number
  name: string
  type: DatasetType
  description: string
  is_public: boolean
  created_at: string
  updated_at: string
  labels?: Label[]
  versions?: DatasetVersion[]
}

export interface DatasetVersion {
  id: number
  dataset_id: number
  version: string
  file_path: string
  checksum: string
  size: number
  created_at: string
}

// ==================== Project Types ====================

export interface Project {
  id: number
  name: string
  description: string
  is_public: boolean
  state: ProjectState
  created_at: string
  updated_at: string
  creator_id: number
  labels?: Label[]
  containers?: Container[]
  datasets?: Dataset[]
  experiment_count?: number
  team_size?: number
}

// ==================== Injection Types ====================

export interface FaultSpec {
  [key: string]: unknown
}

export interface InjectionContainer {
  name: string
  version: string
  namespace: string
}

export interface Injection {
  id: number
  name: string
  project_id: number
  state: InjectionState
  type: InjectionType
  progress: number
  duration: number
  target?: string
  pedestal: InjectionContainer
  benchmark: InjectionContainer
  interval: number
  pre_duration: number
  specs: FaultSpec[][]
  created_at: string
  updated_at: string
  labels?: Label[]
  datapack_id?: string
  groundtruths?: GroundTruth[]
  detector_results?: DetectorResult[]
}

export interface GroundTruth {
  level: 'Service' | 'Pod' | 'Span' | 'Metric'
  target: string
  expected_impact_time?: string
}

export interface DetectorResult {
  span_name: string
  anomaly_type: string
  normal_avg_latency: number
  abnormal_avg_latency: number
  normal_success_rate: number
  abnormal_success_rate: number
  p90?: number
  p95?: number
  p99?: number
}

export interface SubmitInjectionReq {
  project_name: string
  pedestal: InjectionContainer
  benchmark: InjectionContainer
  interval: number
  pre_duration: number
  specs: FaultSpec[][]
  algorithms?: Array<{ name: string; version: string }>
  labels?: Label[]
  name_prefix?: string
}

// ==================== Execution Types ====================

export interface Execution {
  id: number
  algorithm_id: number
  algorithm_version: string
  datapack_id: string
  state: ExecutionState
  execution_duration?: number
  created_at: string
  updated_at: string
  labels?: Label[]
  algorithm?: Container
  datapack?: Datapack
  granularity_results?: GranularityResults
}

export interface Datapack {
  id: string
  injection_id?: number
  dataset_id?: number
  benchmark: InjectionContainer
  created_at: string
}

export interface GranularityResults {
  service_results?: RankResult[]
  pod_results?: RankResult[]
  span_results?: RankResult[]
  metric_results?: RankResult[]
}

export interface RankResult {
  rank: number
  name: string
  confidence: number
  is_ground_truth?: boolean
}

// ==================== Task Types ====================

export enum TaskType {
  SUBMIT_INJECTION = 'SubmitInjection',
  BUILD_DATAPACK = 'BuildDatapack',
  FAULT_INJECTION = 'FaultInjection',
  COLLECT_RESULT = 'CollectResult',
  ALGORITHM_EXECUTION = 'AlgorithmExecution',
}

export interface Task {
  id: string
  type: TaskType
  state: TaskState
  status: number
  trace_id: string
  group_id: string
  parent_id?: string
  project_id?: number
  retry_count: number
  max_retry: number
  immediate: boolean
  created_at: string
  started_at?: string
  finished_at?: string
  payload?: unknown
}

// ==================== User & Auth Types ====================

export interface User {
  id: number
  username: string
  email: string
  full_name?: string
  avatar?: string
  phone?: string
  last_login_at?: string
  status: number
  created_at: string
}

export interface Role {
  id: number
  name: string
  description: string
  is_system: boolean
  created_at: string
  permissions?: Permission[]
}

export interface Permission {
  id: number
  name: string
  action: 'read' | 'write' | 'delete' | 'execute'
  resource_type: string
  description: string
  is_system: boolean
  status: number
}

export interface LoginReq {
  username: string
  password: string
}

export interface LoginRes {
  token: string
  expires_at: string
  user: User
}

// ==================== Evaluation Types ====================

export interface DatapackEvaluationSpec {
  algorithm: { name: string; version: string }
  datapack: string
  filter_labels?: Label[]
}

export interface ExecutionRef {
  execution_id: number
  execution_duration: number
  detector_results: DetectorResult[]
  predictions: GranularityResults
  executed_at: string
}

export interface DatapackEvaluationResult {
  algorithm: string
  algorithm_version: string
  datapack: string
  groundtruths: GroundTruth[]
  execution_refs: ExecutionRef[]
}

// ==================== Fault Injection Types ====================

export interface FaultType {
  id: number
  name: string
  type: string
  category?: string
  description?: string
  parameters?: FaultParameter[]
  created_at?: string
  updated_at?: string
}

export interface FaultParameter {
  name: string
  type: 'string' | 'number' | 'boolean' | 'select' | 'range'
  label: string
  description?: string
  required?: boolean
  default?: any
  options?: string[]
  min?: number
  max?: number
  step?: number
}
