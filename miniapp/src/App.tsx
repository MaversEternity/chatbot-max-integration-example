import { useState, useEffect, useCallback } from 'react';
import { bridge } from './bridge';
import { mockTasks } from './data';
import type { Task } from './types';
import Dashboard from './views/Dashboard';
import TaskDetail from './views/TaskDetail';
import CreateTask from './views/CreateTask';
import styles from './App.module.css';

type View = { type: 'dashboard' } | { type: 'task'; id: number } | { type: 'create' };

function getInitialView(): View {
  // Check URL param (mock) or Bridge start_param (production)
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
    <div className={styles.app}>
      <header className={styles.header}>
        {view.type !== 'dashboard' ? (
          <button className={styles.backBtn} onClick={goBack}>‹ Назад</button>
        ) : <div style={{ width: 60 }} />}
        <div className={styles.title}>{title}</div>
        <div style={{ width: 60 }} />
      </header>
      <main className={styles.content}>
        {view.type === 'dashboard' && (
          <Dashboard tasks={tasks} onOpenTask={id => { bridge.hapticImpact('light'); setView({ type: 'task', id }); }} onCreate={() => { bridge.hapticImpact('medium'); setView({ type: 'create' }); }} />
        )}
        {view.type === 'task' && (
          <TaskDetail task={tasks.find(t => t.id === (view as any).id)!} onSendMessage={(taskId, text) => {
            const now = () => new Date().toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' });
            setTasks(prev => prev.map(t => t.id === taskId ? { ...t, messages: [...t.messages, { from: 'client', text, time: now() }] } : t));
            bridge.hapticImpact('light');
            // Emulate accountant reply
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
          }} />
        )}
        {view.type === 'create' && (
          <CreateTask onSubmit={(name, deadline) => {
            const id = Math.max(...tasks.map(t => t.id)) + 1;
            setTasks(prev => [...prev, { id, name, deadline, status: 'Новая', messages: [], files: [] }]);
            bridge.hapticNotification('success');
            setView({ type: 'dashboard' });
          }} />
        )}
      </main>
    </div>
  );
}
