import { useState } from 'react';
import type { Task } from '../types';
import styles from './TaskDetail.module.css';

interface Props {
  task: Task;
  onSendMessage: (taskId: number, text: string) => void;
}

export default function TaskDetail({ task, onSendMessage }: Props) {
  const [msg, setMsg] = useState('');
  const [showHistory, setShowHistory] = useState(false);
  const [showComments, setShowComments] = useState(true);

  const handleSend = () => {
    if (!msg.trim()) return;
    onSendMessage(task.id, msg.trim());
    setMsg('');
  };

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        <h2 className={styles.title}>{task.name}</h2>

        <div className={styles.fields}>
          <div className={styles.field}>
            <span className={styles.fieldLabel}>Статус</span>
            <span className={`${styles.fieldValue} ${styles.fieldAccent}`}>{task.status}</span>
          </div>
          <div className={styles.field}>
            <span className={styles.fieldLabel}>Срок</span>
            <span className={styles.fieldValue}>
              {task.deadline}
              {deadlineBadge(task.deadline) && (
                <span className={styles.badge}>{deadlineBadge(task.deadline)}</span>
              )}
            </span>
          </div>
          <div className={styles.field}>
            <span className={styles.fieldLabel}>Исполнитель</span>
            <span className={styles.fieldValue}>БухКомпания Финлид</span>
          </div>
          <div className={styles.field}>
            <span className={styles.fieldLabel}>Создана</span>
            <span className={styles.fieldValue}>{formatDate(new Date())}</span>
          </div>
        </div>

        {/* Описание */}
        <div className={styles.section}>
          <div className={styles.sectionTitle}>Описание задачи</div>
          <div className={styles.placeholder}>Укажите детали, если необходимо</div>
        </div>

        {/* Вложения */}
        {task.files.length > 0 && (
          <div className={styles.section}>
            <div className={styles.sectionTitle}>Вложения <span className={styles.countGray}>{task.files.length}</span></div>
            {task.files.map((f, i) => (
              <div key={i} className={styles.fileRow}>
                <span className={styles.fileIcon}>📄</span>
                <span className={styles.fileName}>{f}</span>
              </div>
            ))}
          </div>
        )}

        {/* История */}
        <div className={styles.section}>
          <button className={styles.collapseBtn} onClick={() => setShowHistory(!showHistory)}>
            <span>История</span>
            <span className={styles.chevron}>{showHistory ? '˄' : '˅'}</span>
          </button>
          {showHistory && (
            <div className={styles.historyItem}>Задача создана</div>
          )}
        </div>

        {/* Комментарии */}
        <div className={styles.section}>
          <button className={styles.collapseBtn} onClick={() => setShowComments(!showComments)}>
            <span>{task.messages.length} комментари{suffix(task.messages.length)}</span>
            <span className={styles.chevron}>{showComments ? '˄' : '˅'}</span>
          </button>
          {showComments && (
            <div className={styles.comments}>
              {task.messages.length === 0 && (
                <div className={styles.placeholder}>Добавьте первый комментарий</div>
              )}
              {task.messages.map((m, i) => {
                const isClient = m.from === 'client';
                const name = isClient ? 'Иван Петрович' : 'БухКомпания Финлид';
                return isClient ? (
                  <div key={i} className={styles.clientMsg}>
                    <div className={styles.clientBubble}>{m.text}</div>
                  </div>
                ) : (
                  <div key={i} className={styles.comment}>
                    <div className={styles.avatar}>{initials(name)}</div>
                    <div className={styles.commentBody}>
                      <div className={styles.commentName}>{name}</div>
                      <div className={styles.commentText}>{m.text}</div>
                      <div className={styles.commentTime}>{m.time}</div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      <div className={styles.inputBar}>
        <button className={styles.attachButton}>📎</button>
        <input
          className={styles.input}
          placeholder="Комментарий"
          value={msg}
          onChange={e => setMsg(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleSend()}
        />
        <button className={styles.sendButton} onClick={handleSend}>↑</button>
      </div>
    </div>
  );
}

function formatDate(d: Date) {
  return d.toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' });
}

function initials(name: string) {
  return name.split(' ').map(w => w[0]).slice(0, 2).join('').toUpperCase();
}

function suffix(n: number) {
  if (n === 1) return 'й';
  if (n >= 2 && n <= 4) return 'я';
  return 'ев';
}

function deadlineBadge(deadline: string) {
  const parts = deadline.split('.');
  if (parts.length !== 3) return null;
  const d = new Date(+parts[2], +parts[1] - 1, +parts[0]);
  const diff = Math.ceil((d.getTime() - Date.now()) / 86400000);
  if (diff < 0) return `Просрочено ${-diff} дн.`;
  if (diff === 0) return 'Сегодня';
  if (diff <= 7) return `Через ${diff} дн.`;
  return null;
}
