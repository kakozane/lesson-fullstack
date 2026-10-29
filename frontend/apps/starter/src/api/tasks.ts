import { requestClient } from '#/api/request';

export type TaskStatus = 'todo' | 'doing' | 'done';
export type TaskPriority = 'low' | 'medium' | 'high';
export interface TaskInput {
  title: string;
  description: string;
  assignee: string;
  priority: TaskPriority;
  status: TaskStatus;
  dueDate: string;
}
export interface Task extends TaskInput {
  id: number;
  updatedAt: string;
}
export interface TaskBoard {
  items: Task[];
  version: number;
  total: number;
  doing: number;
  completed: number;
  overdue: number;
  completion: number;
}
export interface TaskMove {
  id: number;
  status: TaskStatus;
  beforeId: number;
}
export const getTaskBoard = () => requestClient.get<TaskBoard>('/tasks');
export const saveTask = (data: TaskInput, version: number, id?: number) =>
  id
    ? requestClient.put<TaskBoard>(`/tasks/${id}`, { ...data, version })
    : requestClient.post<TaskBoard>('/tasks', { ...data, version });
export const removeTask = (id: number, version: number) =>
  requestClient.delete<TaskBoard>(`/tasks/${id}`, { data: { version } });
export const moveTask = ({ id, ...data }: TaskMove, version: number) =>
  requestClient.request<TaskBoard>(`/tasks/${id}/move`, {
    method: 'PATCH',
    data: { ...data, version },
  });
