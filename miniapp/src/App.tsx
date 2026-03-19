import { useState, useEffect, useCallback } from 'react';
import { Button, Container, Flex, Typography } from '@maxhub/max-ui';
import { bridge } from './bridge';
import { mockTasks } from './data';
import type { Task } from './types';
import Dashboard from './views/Dashboard';
import TaskDetail from './views/TaskDetail';
import CreateTask from './views/CreateTask';

type View = { type: 'dashboard' } | { type: 'task'; id: number } | { type: 'create' };

function getInitialView(): View {
  const urlParam = new URLSearchParams(window.location.search).get('startapp');
  const bridgeParam = (window as any).WebApp?.initDataUnsafe?.start_param;
  const param = urlParam || bridgeParam || '';
  const match = param.match(/^task_(\d+)$/);
  if (match) return { type: 'task', id: parseInt(match[1]) };
  return { type: 'dashboard' };
}

export default function App() {
  const [tasks, setTasks] = useState<Task[]>(mockTasks);
  const [view, setView] = useState<View>(getInitialView);

  const goBack = useCallback(() => setView({ type: 'dashboard' }), []);

  useEffect(() => {
    if (view.type !== 'dashboard') {
      bridge.showBackButton();
      bridge.onBackButton(goBack);
      return () => bridge.offBackButton(goBack);
    } else {
      bridge.hideBackButton();
    }
  }, [view, goBack]);

  const title = view.type === 'task'
    ? `#Задача-${(view as any).id}`
    : view.type === 'create' ? 'Новая задача' : 'Финлид';

  return (
    <Flex direction="column" style={{ minHeight: '100vh', background: '#f5f5f7' }}>
      <Flex
        align="center"
        style={{
          padding: '12px 16px', borderBottom: '1px solid #e8e8ec',
          position: 'sticky', top: 0, zIndex: 10, background: '#fff', minHeight: 48
        }}
      >
        <Container style={{ width: 60 }}>
          {view.type !== 'dashboard' && (
            <Button size="small" mode="link" onClick={goBack}>‹ Назад</Button>
          )}
        </Container>
        <Container style={{ flex: 1, textAlign: 'center' }}>
          <Typography.Title variant="small-strong">{title}</Typography.Title>
        </Container>
        <Container style={{ width: 60 }} />
      </Flex>

      <Container style={{ flex: 1 }}>
        {view.type === 'dashboard' && (
          <Dashboard
            tasks={tasks}
            onOpenTask={id => { bridge.hapticImpact('light'); setView({ type: 'task', id }); }}
            onCreate={() => { bridge.hapticImpact('medium'); setView({ type: 'create' }); }}
          />
        )}
        {view.type === 'task' && (
          <TaskDetail
            task={tasks.find(t => t.id === (view as any).id)!}
            onSendMessage={(taskId, text) => {
              const now = () => new Date().toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' });
              setTasks(prev => prev.map(t => t.id === taskId ? { ...t, messages: [...t.messages, { from: 'client', text, time: now() }] } : t));
              bridge.hapticImpact('light');
              setTimeout(() => {
                const replies = [
                  'Добрый день! Приняла в работу, вернусь с ответом в течение часа.',
                  'Спасибо за информацию! Проверю и напишу, если потребуется что-то ещё.',
                  'Получила, спасибо! Документы в порядке.',
                  'Хорошо, учту. Если будут вопросы — напишу.',
                  'Принято! Подготовлю всё к указанному сроку.',
                ];
                const reply = replies[Math.floor(Math.random() * replies.length)];
                setTasks(prev => prev.map(t => t.id === taskId ? { ...t, messages: [...t.messages, { from: 'accountant', text: reply, time: now() }] } : t));
                bridge.hapticNotification('success');
              }, 2000);
            }}
          />
        )}
        {view.type === 'create' && (
          <CreateTask onSubmit={(name, deadline) => {
            const id = Math.max(...tasks.map(t => t.id)) + 1;
            setTasks(prev => [...prev, { id, name, deadline, status: 'Новая', messages: [], files: [] }]);
            bridge.hapticNotification('success');
            setView({ type: 'dashboard' });
          }} />
        )}
      </Container>
    </Flex>
  );
}
