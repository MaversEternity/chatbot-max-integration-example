export interface Task {
  id: number;
  name: string;
  deadline: string;
  status: TaskStatus;
  messages: TaskMessage[];
  files: string[];
}

export type TaskStatus = 'Новая' | 'В работе' | 'На проверке' | 'Завершена' | 'Просрочено';

export interface TaskMessage {
  from: 'accountant' | 'client';
  text: string;
  time: string;
}
