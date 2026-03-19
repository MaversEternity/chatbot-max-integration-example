import { useState } from 'react';
import {
  Button, CellHeader, CellList, CellSimple, Container, Counter,
  Flex, Grid, Panel, SearchInput, Typography
} from '@maxhub/max-ui';
import { bridge } from '../bridge';
import type { Task, TaskStatus } from '../types';

interface Props {
  tasks: Task[];
  onOpenTask: (id: number) => void;
  onCreate: () => void;
}

const filters = [
  { key: 'active', label: 'Актуальные' },
  { key: 'new', label: 'Новые' },
  { key: 'progress', label: 'В работе' },
  { key: 'review', label: 'На проверке' },
  { key: 'done', label: 'Завершённые' },
  { key: 'all', label: 'Все' },
] as const;

const statusColor: Record<TaskStatus, string> = {
  'Просрочено': '#ff5252',
  'В работе': '#ffa726',
  'Новая': '#4a7cff',
  'На проверке': '#ab47bc',
  'Завершена': '#4caf50',
};

const statusEmoji: Record<TaskStatus, string> = {
  'Просрочено': '🔴',
  'В работе': '🟡',
  'Новая': '🔵',
  'На проверке': '🟣',
  'Завершена': '🟢',
};

function matchesFilter(task: Task, filter: string): boolean {
  switch (filter) {
    case 'active': return task.status !== 'Завершена';
    case 'overdue': return task.status === 'Просрочено';
    case 'new': return task.status === 'Новая';
    case 'progress': return task.status === 'В работе';
    case 'review': return task.status === 'На проверке';
    case 'done': return task.status === 'Завершена';
    case 'all': return true;
    default: return true;
  }
}

export default function Dashboard({ tasks, onOpenTask, onCreate }: Props) {
  const [filter, setFilter] = useState('active');
  const [search, setSearch] = useState('');
  const user = bridge.user;

  const active = tasks.filter(t => ['В работе', 'Новая', 'На проверке'].includes(t.status)).length;
  const overdue = tasks.filter(t => t.status === 'Просрочено').length;
  const done = tasks.filter(t => t.status === 'Завершена').length;

  let filtered = tasks.filter(t => matchesFilter(t, filter));
  if (search) {
    const q = search.toLowerCase();
    filtered = filtered.filter(t => t.name.toLowerCase().includes(q) || `#${t.id}`.includes(q));
  }

  return (
    <Container style={{ padding: 16, paddingBottom: 80 }}>
      <Container style={{ marginBottom: 16 }}>
        <Typography.Body>Добрый день,</Typography.Body>
        <Typography.Title>{user.first_name} {user.last_name}</Typography.Title>
      </Container>

      <Grid cols={3} gap={8} style={{ marginBottom: 16 }}>
        <Panel mode="secondary" centeredX style={{ padding: '14px 8px', cursor: 'pointer' }}
          onClick={() => { setFilter('active'); bridge.hapticSelection(); }}>
          <Typography.Display style={{ color: '#5b9aff' }}>{active}</Typography.Display>
          <Typography.Label>В работе</Typography.Label>
        </Panel>
        <Panel mode="secondary" centeredX style={{ padding: '14px 8px', cursor: 'pointer' }}
          onClick={() => { setFilter('overdue'); bridge.hapticSelection(); }}>
          <Typography.Display style={{ color: '#ff6b6b' }}>{overdue}</Typography.Display>
          <Typography.Label>Просрочено</Typography.Label>
        </Panel>
        <Panel mode="secondary" centeredX style={{ padding: '14px 8px', cursor: 'pointer' }}
          onClick={() => { setFilter('done'); bridge.hapticSelection(); }}>
          <Typography.Display style={{ color: '#66d9a0' }}>{done}</Typography.Display>
          <Typography.Label>Выполнено</Typography.Label>
        </Panel>
      </Grid>

      <Container style={{ marginBottom: 12 }}>
        <SearchInput
          placeholder="Поиск задач..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </Container>

      <Flex gap={6} className="filters-scroll" style={{ overflowX: 'auto', paddingBottom: 12 }}>
        {filters.map(f => (
          <Button
            key={f.key}
            size="small"
            mode={filter === f.key ? 'primary' : 'secondary'}
            onClick={() => { setFilter(f.key); bridge.hapticSelection(); }}
          >
            {f.label}
          </Button>
        ))}
      </Flex>

      <CellList mode="island" header={<CellHeader>Задачи</CellHeader>}>
        {filtered.length === 0 && (
          <CellSimple title="Задач не найдено" />
        )}
        {filtered.map(task => (
          <CellSimple
            key={task.id}
            title={task.name}
            subtitle={
              <span>
                {statusEmoji[task.status]} {task.status} · 📅 {task.deadline}
                {task.files.length > 0 && ` · 📎 ${task.files.length}`}
                {task.messages.length > 0 && ` · 💬 ${task.messages.length}`}
              </span>
            }
            overline={`#${task.id}`}
            before={
              <Container style={{
                width: 4, height: '100%', minHeight: 32,
                borderRadius: 2, background: statusColor[task.status]
              }} />
            }
            after={task.messages.length > 0 && task.messages[task.messages.length - 1].from === 'accountant'
              ? <Counter value={task.messages.length}>{task.messages.length}</Counter>
              : undefined
            }
            showChevron
            onClick={() => onOpenTask(task.id)}
          />
        ))}
      </CellList>

      <Container style={{ position: 'fixed', bottom: 16, left: 16, right: 16, zIndex: 5 }}>
        <Button size="large" mode="primary" stretched onClick={onCreate}>
          + Создать задачу
        </Button>
      </Container>
    </Container>
  );
}
