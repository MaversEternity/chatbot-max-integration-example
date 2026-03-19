import { useState } from 'react';
import { Button, CellHeader, CellList, Container, Flex, Input, Textarea, Typography } from '@maxhub/max-ui';

interface Props {
  onSubmit: (name: string, deadline: string) => void;
}

export default function CreateTask({ onSubmit }: Props) {
  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');
  const [date, setDate] = useState('');
  const [priority, setPriority] = useState<'low' | 'med' | 'high'>('med');

  const handleSubmit = () => {
    if (!name.trim() || !date.trim()) return;
    onSubmit(name.trim(), date.trim());
  };

  return (
    <Container style={{ padding: 16, minHeight: '100vh' }}>
      <Typography.Title>Создать задачу</Typography.Title>

      <CellList mode="island" header={<CellHeader>Название</CellHeader>}>
        <Input
          placeholder="Что нужно сделать?"
          value={name}
          onChange={e => setName(e.target.value)}
        />
      </CellList>

      <CellList mode="island" header={<CellHeader>Описание</CellHeader>}>
        <Textarea
          placeholder="Подробности..."
          value={desc}
          onChange={e => setDesc(e.target.value)}
        />
      </CellList>

      <CellList mode="island" header={<CellHeader>Срок выполнения</CellHeader>}>
        <Input
          placeholder="дд.мм.гггг"
          value={date}
          onChange={e => setDate(e.target.value)}
        />
      </CellList>

      <CellList mode="island" header={<CellHeader>Приоритет</CellHeader>}>
        <Flex gap={8} style={{ padding: 8 }}>
          <Button
            size="small"
            mode={priority === 'low' ? 'primary' : 'secondary'}
            appearance={priority === 'low' ? 'neutral-themed' : 'neutral'}
            onClick={() => setPriority('low')}
          >
            Низкий
          </Button>
          <Button
            size="small"
            mode={priority === 'med' ? 'primary' : 'secondary'}
            onClick={() => setPriority('med')}
          >
            Средний
          </Button>
          <Button
            size="small"
            mode={priority === 'high' ? 'primary' : 'secondary'}
            appearance={priority === 'high' ? 'negative' : 'neutral'}
            onClick={() => setPriority('high')}
          >
            Высокий
          </Button>
        </Flex>
      </CellList>

      <Container style={{ marginTop: 24 }}>
        <Button size="large" mode="primary" stretched onClick={handleSubmit}>
          Создать задачу
        </Button>
      </Container>
    </Container>
  );
}
