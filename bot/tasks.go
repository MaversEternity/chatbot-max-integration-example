package bot

import (
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusNew      TaskStatus = "Новая"
	StatusProgress TaskStatus = "В работе"
	StatusReview   TaskStatus = "На проверке"
	StatusDone     TaskStatus = "Завершена"
	StatusOverdue  TaskStatus = "Просрочено"
)

type TaskMessage struct {
	From      string    `json:"from"`
	Text      string    `json:"text"`
	Filename  string    `json:"filename,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskFile struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Task struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Deadline  string        `json:"deadline"`
	Status    TaskStatus    `json:"status"`
	Priority  string        `json:"priority,omitempty"`
	Messages  []TaskMessage `json:"messages"`
	Files     []TaskFile    `json:"files"`
	CreatedAt time.Time     `json:"created_at"`
}

type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]*Task
	nextID int
}

func NewTaskStore() *TaskStore {
	s := &TaskStore{
		tasks:  make(map[int]*Task),
		nextID: 46,
	}
	s.seed()
	return s
}

func (s *TaskStore) seed() {
	seeds := []*Task{
		{ID: 38, Name: "Сдать отчёт по НДС", Deadline: "10.03.2026", Status: StatusOverdue, CreatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{ID: 40, Name: "Закрывающие документы за февраль", Deadline: "25.03.2026", Status: StatusReview, Files: []TaskFile{{Name: "Акт_сверки.pdf", URL: "https://example.com/files/40/Акт_сверки.pdf"}, {Name: "Накладная_01.pdf", URL: "https://example.com/files/40/Накладная_01.pdf"}, {Name: "Накладная_02.pdf", URL: "https://example.com/files/40/Накладная_02.pdf"}, {Name: "УПД_февраль.pdf", URL: "https://example.com/files/40/УПД_февраль.pdf"}}, CreatedAt: time.Now().Add(-14 * 24 * time.Hour)},
		{ID: 42, Name: "Подготовить декларацию за 1 кв.", Deadline: "15.04.2026", Status: StatusProgress, Files: []TaskFile{{Name: "Выписка_Сбер_Q1_2026.pdf", URL: "https://example.com/files/42/Выписка_Сбер_Q1_2026.pdf"}}, CreatedAt: time.Now().Add(-2 * 24 * time.Hour)},
		{ID: 43, Name: "Сверка с контрагентом ООО «Альфа»", Deadline: "20.04.2026", Status: StatusNew, CreatedAt: time.Now().Add(-1 * 24 * time.Hour)},
		{ID: 44, Name: "Начисление зарплаты за март", Deadline: "05.04.2026", Status: StatusProgress, CreatedAt: time.Now().Add(-3 * 24 * time.Hour)},
		{ID: 45, Name: "Регистрация ККТ", Deadline: "30.04.2026", Status: StatusNew, CreatedAt: time.Now()},
	}
	for _, t := range seeds {
		if t.Messages == nil {
			t.Messages = []TaskMessage{}
		}
		if t.Files == nil {
			t.Files = []TaskFile{}
		}
		s.tasks[t.ID] = t
	}
	// Add messages to task 42
	s.tasks[42].Messages = []TaskMessage{
		{From: "accountant", Text: "Добрый день! Пришлите, пожалуйста, выписку из банка за январь–март и акты сверки.", Timestamp: time.Now().Add(-1 * time.Hour)},
	}
	s.tasks[38].Messages = []TaskMessage{
		{From: "accountant", Text: "Срочно нужны документы для сдачи отчёта.", Timestamp: time.Now().Add(-48 * time.Hour)},
		{From: "client", Text: "Подготовлю к вечеру", Timestamp: time.Now().Add(-47 * time.Hour)},
		{From: "accountant", Text: "Жду, срок уже прошёл.", Timestamp: time.Now().Add(-24 * time.Hour)},
	}
}

func (s *TaskStore) Create(name, deadline string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &Task{
		ID:        s.nextID,
		Name:      name,
		Deadline:  deadline,
		Status:    StatusNew,
		Messages:  []TaskMessage{},
		Files:     []TaskFile{},
		CreatedAt: time.Now(),
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t
}

func (s *TaskStore) Get(id int) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *TaskStore) ListActive() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Task
	for _, t := range s.tasks {
		if t.Status != StatusDone {
			result = append(result, t)
		}
	}
	return result
}

func (s *TaskStore) ListAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Task
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *TaskStore) ListDone() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Task
	for _, t := range s.tasks {
		if t.Status == StatusDone {
			result = append(result, t)
		}
	}
	return result
}

func (s *TaskStore) AddMessage(taskID int, msg TaskMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tasks[taskID]; ok {
		t.Messages = append(t.Messages, msg)
	}
}

func (s *TaskStore) AddFile(taskID int, filename, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tasks[taskID]; ok {
		t.Files = append(t.Files, TaskFile{Name: filename, URL: url})
	}
}
