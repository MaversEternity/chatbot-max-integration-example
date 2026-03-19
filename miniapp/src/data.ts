import type { Task } from './types';

// Replace with API calls to your backend in production
export const mockTasks: Task[] = [
  {
    id: 38, name: 'Сдать отчёт по НДС', deadline: '10.03.2026', status: 'Просрочено',
    files: ['Декларация_НДС.pdf', 'Книга_продаж.xlsx'],
    messages: [
      { from: 'accountant', text: 'Срочно нужны документы для сдачи отчёта.', time: '15:00' },
      { from: 'client', text: 'Подготовлю к вечеру', time: '15:30' },
      { from: 'accountant', text: 'Жду, срок уже прошёл.', time: '09:00' },
    ],
  },
  {
    id: 40, name: 'Закрывающие документы за февраль', deadline: '25.03.2026', status: 'На проверке',
    files: ['Акт_сверки.pdf', 'Накладная_01.pdf', 'Накладная_02.pdf', 'УПД_февраль.pdf'],
    messages: [],
  },
  {
    id: 42, name: 'Подготовить декларацию за 1 кв.', deadline: '15.04.2026', status: 'В работе',
    files: ['Выписка_Сбер_Q1_2026.pdf'],
    messages: [
      { from: 'accountant', text: 'Добрый день! Для подготовки декларации потребуются:\n1. Выписка из банка за янв–март\n2. Акты сверки\n3. Счета-фактуры', time: '10:30' },
      { from: 'client', text: 'Здравствуйте! Выписку отправляю, акты будут завтра.', time: '10:45' },
      { from: 'accountant', text: 'Получила, спасибо! Жду акты сверки.', time: '10:48' },
      { from: 'client', text: 'Вот акты и счета-фактуры', time: '09:15' },
      { from: 'accountant', text: 'Все документы получены! Приступаю к подготовке. Ориентировочно к 10 апреля.', time: '09:30' },
    ],
  },
  {
    id: 43, name: 'Сверка с контрагентом ООО «Альфа»', deadline: '20.04.2026', status: 'Новая',
    files: [], messages: [],
  },
  {
    id: 44, name: 'Начисление зарплаты за март', deadline: '05.04.2026', status: 'В работе',
    files: [],
    messages: [{ from: 'accountant', text: 'Пришлите табель учёта рабочего времени за март.', time: '11:00' }],
  },
  {
    id: 45, name: 'Регистрация ККТ', deadline: '30.04.2026', status: 'Новая',
    files: [], messages: [],
  },
];
