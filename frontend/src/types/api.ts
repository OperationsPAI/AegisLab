
export enum InjectionState {
  PENDING = '0',
  RUNNING = '1',
  COMPLETED = '2',
  ERROR = '3',
  STOPPED = '4',
}


export enum InjectionType {
  NETWORK = 'network',
  CPU = 'cpu',
  MEMORY = 'memory',
  DISK = 'disk',
  PROCESS = 'process',
  KUBERNETES = 'kubernetes',
}


export enum ProjectState {
  ACTIVE = 0,
  PAUSED = 1,
  COMPLETED = 2,
  ARCHIVED = 3,
}


export interface FaultType {
  id: number;
  name: string;
  type: string;
  category?: string;
  description?: string;
  parameters?: FaultParameter[];
  created_at?: string;
  updated_at?: string;
}


export interface FaultParameter {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'range';
  label: string;
  description?: string;
  required?: boolean;
  default?: string | number | boolean;
  options?: string[];
  min?: number;
  max?: number;
  step?: number;
}
