import { useState } from 'react';
import {
  Avatar, Button, CellAction, CellHeader, CellList, CellSimple,
  Container, Flex, IconButton, Input, Panel, Typography
} from '@maxhub/max-ui';
import type { Task } from '../types';

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
    <Container style={{ minHeight: '100vh', paddingBottom: 70 }}>
      <Panel mode="primary" style={{ margin: 12, borderRadius: 16, padding: 20 }}>
        <Typography.Headline variant="medium-strong">{task.name}</Typography.Headline>

        <CellList style={{ marginTop: 16 }}>
          <CellSimple
            height="compact"
            title={<Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>Статус</Typography.Body>}
            after={<Typography.Body variant="small-strong" style={{ color: 'var(--max-accent, #7c5ce7)' }}>{task.status}</Typography.Body>}
          />
          <CellSimple
            height="compact"
            title={<Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>Срок</Typography.Body>}
            after={
              <Flex align="center" gap={8}>
                <Typography.Body variant="small-strong">{task.deadline}</Typography.Body>
                {deadlineBadge(task.deadline) && (
                  <Typography.Label variant="small-caps" style={{
                    background: 'var(--max-bg_tertiary, #333)', color: '#7c5ce7',
                    padding: '2px 8px', borderRadius: 10
                  }}>
                    {deadlineBadge(task.deadline)}
                  </Typography.Label>
                )}
              </Flex>
            }
          />
          <CellSimple
            height="compact"
            title={<Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>Исполнитель</Typography.Body>}
            after={<Typography.Body variant="small-strong">БухКомпания Финлид</Typography.Body>}
          />
          <CellSimple
            height="compact"
            title={<Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>Создана</Typography.Body>}
            after={<Typography.Body variant="small-strong">{formatDate(new Date())}</Typography.Body>}
          />
        </CellList>

        {/* Описание */}
        <CellList style={{ marginTop: 16 }} header={<CellHeader>Описание задачи</CellHeader>}>
          <CellSimple title={<Typography.Body variant="small" style={{ color: 'var(--max-text_tertiary, #b0b0bc)' }}>Укажите детали, если необходимо</Typography.Body>} />
        </CellList>

        {/* Вложения */}
        {task.files.length > 0 && (
          <CellList
            style={{ marginTop: 16 }}
            header={
              <CellHeader after={<Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>{task.files.length}</Typography.Body>}>
                Вложения
              </CellHeader>
            }
          >
            {task.files.map((f, i) => (
              <CellSimple key={i} title={f} before={<span style={{ fontSize: 18 }}>📄</span>} />
            ))}
          </CellList>
        )}

        {/* История */}
        <Container style={{ marginTop: 16, borderTop: '1px solid var(--max-bg_tertiary, #333)', paddingTop: 16 }}>
          <CellAction onClick={() => setShowHistory(!showHistory)}>
            <Flex justify="space-between" align="center" style={{ width: '100%' }}>
              <Typography.Body variant="small-strong">История</Typography.Body>
              <Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>{showHistory ? '˄' : '˅'}</Typography.Body>
            </Flex>
          </CellAction>
          {showHistory && (
            <Typography.Body variant="small" style={{ color: '#8e8e9a', padding: '4px 16px 8px' }}>
              Задача создана
            </Typography.Body>
          )}
        </Container>

        {/* Комментарии */}
        <Container style={{ marginTop: 8, borderTop: '1px solid var(--max-bg_tertiary, #333)', paddingTop: 16 }}>
          <CellAction onClick={() => setShowComments(!showComments)}>
            <Flex justify="space-between" align="center" style={{ width: '100%' }}>
              <Typography.Body variant="small-strong">
                {task.messages.length} комментари{suffix(task.messages.length)}
              </Typography.Body>
              <Typography.Body variant="small" style={{ color: 'var(--max-text_secondary, #8e8e9a)' }}>{showComments ? '˄' : '˅'}</Typography.Body>
            </Flex>
          </CellAction>

          {showComments && (
            <Flex direction="column" gap={16} style={{ padding: '8px 0' }}>
              {task.messages.length === 0 && (
                <Typography.Body variant="small" style={{ color: 'var(--max-text_tertiary, #b0b0bc)' }}>
                  Добавьте первый комментарий
                </Typography.Body>
              )}
              {task.messages.map((m, i) => {
                const isClient = m.from === 'client';
                const name = isClient ? 'Иван Петрович' : 'БухКомпания Финлид';
                const grad = isClient ? 'green' as const : 'purple' as const;

                return isClient ? (
                  <Flex key={i} justify="end">
                    <Panel mode="secondary" style={{
                      padding: '10px 14px', borderRadius: '16px 16px 4px 16px',
                      maxWidth: '80%', background: 'var(--max-bg_tertiary, #333)'
                    }}>
                      <Typography.Body variant="small">{m.text}</Typography.Body>
                    </Panel>
                  </Flex>
                ) : (
                  <Flex key={i} gap={10}>
                    <Avatar.Container size={36}>
                      <Avatar.Text gradient={grad}>{initials(name)}</Avatar.Text>
                    </Avatar.Container>
                    <Flex direction="column" style={{ flex: 1, minWidth: 0 }}>
                      <Typography.Label variant="small-strong">{name}</Typography.Label>
                      <Typography.Body variant="small" style={{ lineHeight: 1.45, whiteSpace: 'pre-wrap' }}>
                        {m.text}
                      </Typography.Body>
                      <Typography.Label variant="small-caps" style={{ color: '#b0b0bc', marginTop: 4 }}>
                        {m.time}
                      </Typography.Label>
                    </Flex>
                  </Flex>
                );
              })}
            </Flex>
          )}
        </Container>
      </Panel>

      {/* Bottom input bar */}
      <Flex
        align="center"
        gap={10}
        style={{
          position: 'fixed', bottom: 0, left: 0, right: 0,
          padding: '10px 16px', background: 'var(--max-bg_secondary, #222)',
          borderTop: '1px solid var(--max-bg_tertiary, #333)'
        }}
      >
        <IconButton size="medium" mode="tertiary">📎</IconButton>
        <Container style={{ flex: 1 }}>
          <Input
            placeholder="Комментарий"
            value={msg}
            onChange={e => setMsg(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleSend()}
          />
        </Container>
        <Button size="small" mode="primary" onClick={handleSend}>↑</Button>
      </Flex>
    </Container>
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
