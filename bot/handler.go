package bot

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type ConvState int

const (
	StateIdle ConvState = iota
	StateAwaitingTaskName
	StateAwaitingTaskDate
	StateAwaitingReply
	StateAwaitingFile
)

type UserState struct {
	State       ConvState
	TaskName    string
	ReplyTaskID int
}

type Handler struct {
	api    *maxbot.Api
	tasks  *TaskStore
	mu     sync.RWMutex
	states map[int64]*UserState
}

func NewHandler(api *maxbot.Api, tasks *TaskStore) *Handler {
	return &Handler{
		api:    api,
		tasks:  tasks,
		states: make(map[int64]*UserState),
	}
}

func (h *Handler) getState(userID int64) *UserState {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.states[userID]
	if !ok {
		s = &UserState{State: StateIdle}
		h.states[userID] = s
	}
	return s
}

func (h *Handler) Run(ctx context.Context) {
	log.Println("Handler.Run: waiting for updates on channel...")
	for upd := range h.api.GetUpdates(ctx) {
		log.Printf("Handler.Run: received update type=%T", upd)
		switch u := upd.(type) {
		case *schemes.MessageCreatedUpdate:
			h.handleMessage(ctx, u)
		case *schemes.MessageCallbackUpdate:
			h.handleCallback(ctx, u)
		case *schemes.BotStartedUpdate:
			h.sendWelcome(ctx, u.ChatId)
		}
	}
}

func (h *Handler) handleMessage(ctx context.Context, u *schemes.MessageCreatedUpdate) {
	chatID := u.Message.Recipient.ChatId
	userID := u.Message.Sender.UserId
	text := strings.TrimSpace(u.Message.Body.Text)
	state := h.getState(userID)

	switch state.State {
	case StateAwaitingTaskName:
		state.TaskName = text
		state.State = StateAwaitingTaskDate
		h.send(ctx, chatID, "Укажите срок выполнения (дд.мм.гггг):")

	case StateAwaitingTaskDate:
		task := h.tasks.Create(state.TaskName, text)
		state.State = StateIdle
		state.TaskName = ""
		msg := fmt.Sprintf("✅ *Задача создана*\n\n📋 *#%d %s*\n📅 Срок: %s\n📊 Статус: %s",
			task.ID, task.Name, task.Deadline, task.Status)
		h.sendWithKeyboard(ctx, chatID, msg, h.taskCreatedKeyboard(task.ID))

	case StateAwaitingFile:
		// In real MAX, u.Message.Body would contain file attachments.
		// Here we simulate by treating the text as a filename.
		filename := text
		if filename == "" {
			filename = "document.pdf"
		}
		// In production, upload to S3 and use real URL. For demo, use placeholder.
		fakeURL := fmt.Sprintf("https://example.com/files/%d/%s", state.ReplyTaskID, filename)
		h.tasks.AddFile(state.ReplyTaskID, filename, fakeURL)
		h.tasks.AddMessage(state.ReplyTaskID, TaskMessage{
			From:      "client",
			Text:      fmt.Sprintf("📎 Прикреплён файл: %s", filename),
			Filename:  filename,
			Timestamp: time.Now(),
		})
		task := h.tasks.Get(state.ReplyTaskID)
		state.State = StateIdle
		var taskName string
		if task != nil {
			taskName = task.Name
		}
		msg := fmt.Sprintf("✅ Файл *%s* прикреплён к задаче *#%d*\n_%s_\n📎 Всего файлов: %d",
			filename, state.ReplyTaskID, taskName, len(task.Files))
		h.sendWithKeyboard(ctx, chatID, msg, h.taskDetailKeyboard(state.ReplyTaskID))
		state.ReplyTaskID = 0

	case StateAwaitingReply:
		h.tasks.AddMessage(state.ReplyTaskID, TaskMessage{
			From:      "client",
			Text:      text,
			Timestamp: time.Now(),
		})
		task := h.tasks.Get(state.ReplyTaskID)
		state.State = StateIdle
		var taskName string
		if task != nil {
			taskName = task.Name
		}
		msg := fmt.Sprintf("✅ Сообщение доставлено по задаче *#%d*\n_%s_", state.ReplyTaskID, taskName)
		h.sendWithKeyboard(ctx, chatID, msg, h.mainMenuKeyboard())
		state.ReplyTaskID = 0

	default:
		// Handle text commands
		switch text {
		case "/start":
			h.sendWelcome(ctx, chatID)
		default:
			// Free-form chat message → forward to accountant, always show menu
			h.sendWithKeyboard(ctx, chatID, fmt.Sprintf("✅ Сообщение доставлено бухгалтеру:\n_%s_", text), h.mainMenuKeyboard())
			// Accountant auto-replies disabled for now
		}
	}
}

func (h *Handler) handleCallback(ctx context.Context, u *schemes.MessageCallbackUpdate) {
	payload := u.Callback.Payload
	userID := u.Callback.User.UserId
	state := h.getState(userID)

	var chatID int64
	if u.Message != nil {
		chatID = u.Message.Recipient.ChatId
	}

	// Answer callback to remove loading state
	h.api.Messages.AnswerOnCallback(ctx, u.Callback.CallbackID, &schemes.CallbackAnswer{})

	switch {
	case payload == "noop":
		return

	case payload == "create_task":
		state.State = StateAwaitingTaskName
		h.send(ctx, chatID, "Введите название задачи:")

	case payload == "my_tasks" || payload == "tasks_filter:active":
		h.sendTaskList(ctx, chatID, "active", 0)

	case payload == "tasks_filter:done":
		h.sendTaskList(ctx, chatID, "done", 0)

	case payload == "tasks_filter:all":
		h.sendTaskList(ctx, chatID, "all", 0)

	case strings.HasPrefix(payload, "tasks_page:"):
		// format: tasks_page:filter:offset
		parts := strings.SplitN(strings.TrimPrefix(payload, "tasks_page:"), ":", 2)
		if len(parts) == 2 {
			offset, _ := strconv.Atoi(parts[1])
			h.sendTaskList(ctx, chatID, parts[0], offset)
		}

	case payload == "open_cabinet":
		// In production, this is an open_app button that opens directly.
		// In mock mode, we just acknowledge — the client simulator handles the redirect.
		return

	case strings.HasPrefix(payload, "reply_task:"):
		idStr := strings.TrimPrefix(payload, "reply_task:")
		id, _ := strconv.Atoi(idStr)
		task := h.tasks.Get(id)
		if task != nil {
			state.State = StateAwaitingReply
			state.ReplyTaskID = id
			h.send(ctx, chatID, fmt.Sprintf("💬 Отвечаете по задаче *#%d %s*\n\nВведите сообщение:", task.ID, task.Name))
		}

	case strings.HasPrefix(payload, "attach_task:"):
		idStr := strings.TrimPrefix(payload, "attach_task:")
		id, _ := strconv.Atoi(idStr)
		task := h.tasks.Get(id)
		if task != nil {
			state.State = StateAwaitingFile
			state.ReplyTaskID = id
			h.send(ctx, chatID, fmt.Sprintf("📎 Прикрепление файла к задаче *#%d %s*\n\nОтправьте файл или введите название файла:", task.ID, task.Name))
		}

	case strings.HasPrefix(payload, "task_files:"):
		idStr := strings.TrimPrefix(payload, "task_files:")
		id, _ := strconv.Atoi(idStr)
		task := h.tasks.Get(id)
		if task != nil && len(task.Files) > 0 {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("📄 *Файлы задачи #%d* (%d):\n", task.ID, len(task.Files)))
			for i, f := range task.Files {
				sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, f.Name))
			}
			kb := &maxbot.Keyboard{}
			for _, f := range task.Files {
				if f.URL != "" {
					kb.AddRow().AddLink("📄 "+f.Name, schemes.DEFAULT, f.URL)
				}
			}
			kb.AddRow().AddCallback("⬅️ К задаче", schemes.DEFAULT, fmt.Sprintf("open_task:%d", id))
			h.sendWithKeyboard(ctx, chatID, sb.String(), kb)
		}

	case strings.HasPrefix(payload, "open_task:"):
		idStr := strings.TrimPrefix(payload, "open_task:")
		id, _ := strconv.Atoi(idStr)
		task := h.tasks.Get(id)
		if task != nil {
			h.sendTaskDetail(ctx, chatID, task)
		}
	}
}

func (h *Handler) sendWelcome(ctx context.Context, chatID int64) {
	msg := "Добро пожаловать! Я бот бухгалтерской компании *Финлид*.\n\nВы можете написать сообщение — оно будет доставлено вашему бухгалтеру. Или воспользуйтесь кнопками ниже."
	h.sendWithKeyboard(ctx, chatID, msg, h.mainMenuKeyboard())
}

const tasksPerPage = 5

func (h *Handler) sendTaskList(ctx context.Context, chatID int64, filter string, offset int) {
	var tasks []*Task
	var title string
	switch filter {
	case "active":
		tasks = h.tasks.ListActive()
		title = "Ваши актуальные задачи:"
	case "done":
		tasks = h.tasks.ListDone()
		title = "Завершённые задачи:"
	default:
		tasks = h.tasks.ListAll()
		title = "Все задачи:"
	}

	if len(tasks) == 0 {
		h.sendWithKeyboard(ctx, chatID, "Задач не найдено.", h.taskListFilterKeyboard())
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	total := len(tasks)
	if offset >= total {
		offset = 0
	}
	end := offset + tasksPerPage
	if end > total {
		end = total
	}
	page := tasks[offset:end]
	pageNum := (offset / tasksPerPage) + 1
	totalPages := (total + tasksPerPage - 1) / tasksPerPage

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%s* (стр. %d/%d, всего: %d)\n", title, pageNum, totalPages, total))
	for _, t := range page {
		emoji := statusEmoji(t.Status)
		sb.WriteString(fmt.Sprintf("\n%s *#%d* %s\n      📅 %s · *%s*",
			emoji, t.ID, t.Name, t.Deadline, t.Status))
		if len(t.Messages) > 0 {
			sb.WriteString(fmt.Sprintf(" · 💬 %d", len(t.Messages)))
		}
		if len(t.Files) > 0 {
			sb.WriteString(fmt.Sprintf(" · 📎 %d", len(t.Files)))
		}
	}
	sb.WriteString("\n\nВыберите задачу:")

	kb := &maxbot.Keyboard{}
	for _, t := range page {
		emoji := statusEmoji(t.Status)
		kb.AddRow().AddCallback(fmt.Sprintf("%s #%d %s", emoji, t.ID, t.Name), schemes.DEFAULT, fmt.Sprintf("open_task:%d", t.ID))
	}

	// Pagination row
	if totalPages > 1 {
		navRow := kb.AddRow()
		if offset > 0 {
			navRow.AddCallback("⬅️ Назад", schemes.DEFAULT, fmt.Sprintf("tasks_page:%s:%d", filter, offset-tasksPerPage))
		}
		navRow.AddCallback(fmt.Sprintf("%d/%d", pageNum, totalPages), schemes.DEFAULT, "noop")
		if end < total {
			navRow.AddCallback("Вперёд ➡️", schemes.DEFAULT, fmt.Sprintf("tasks_page:%s:%d", filter, end))
		}
	}

	// Filter row
	row := kb.AddRow()
	row.AddCallback("Актуальные", schemes.DEFAULT, "tasks_filter:active")
	row.AddCallback("Завершённые", schemes.DEFAULT, "tasks_filter:done")
	row.AddCallback("Все", schemes.DEFAULT, "tasks_filter:all")
	h.sendWithKeyboard(ctx, chatID, sb.String(), kb)
}

func (h *Handler) sendTaskDetail(ctx context.Context, chatID int64, t *Task) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 *Задача #%d*\n*%s*\n\n📅 Срок: %s\n📊 Статус: %s",
		t.ID, t.Name, t.Deadline, t.Status))

	if len(t.Files) > 0 {
		sb.WriteString(fmt.Sprintf("\n\n📎 *Файлы (%d):*", len(t.Files)))
		for i, f := range t.Files {
			sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, f.Name))
		}
	}

	if len(t.Messages) > 0 {
		sb.WriteString("\n\n*Последние сообщения:*")
		start := 0
		if len(t.Messages) > 3 {
			start = len(t.Messages) - 3
		}
		for _, m := range t.Messages[start:] {
			from := "Бухгалтер"
			if m.From == "client" {
				from = "Вы"
			}
			sb.WriteString(fmt.Sprintf("\n_%s:_ %s", from, m.Text))
		}
	}

	kb := h.taskDetailKeyboard(t.ID)
	if len(t.Files) > 0 {
		kb.AddRow().AddCallback(fmt.Sprintf("📄 Файлы (%d)", len(t.Files)), schemes.DEFAULT, fmt.Sprintf("task_files:%d", t.ID))
	}
	kb.AddRow().AddCallback("📑 К списку", schemes.DEFAULT, "my_tasks")
	h.sendWithKeyboard(ctx, chatID, sb.String(), kb)
}

func (h *Handler) send(ctx context.Context, chatID int64, text string) {
	err := h.api.Messages.Send(ctx, maxbot.NewMessage().SetChat(chatID).SetText(text).SetFormat("markdown"))
	if err != nil {
		log.Printf("ERROR send to chat %d: %v", chatID, err)
	} else {
		log.Printf("OK sent to chat %d: %s", chatID, text[:min(len(text), 50)])
	}
}

func (h *Handler) sendWithKeyboard(ctx context.Context, chatID int64, text string, kb *maxbot.Keyboard) {
	err := h.api.Messages.Send(ctx, maxbot.NewMessage().SetChat(chatID).SetText(text).SetFormat("markdown").AddKeyboard(kb))
	if err != nil {
		log.Printf("ERROR sendWithKeyboard to chat %d: %v", chatID, err)
	} else {
		log.Printf("OK sent with keyboard to chat %d: %s", chatID, text[:min(len(text), 50)])
	}
}

// Keyboards

func (h *Handler) mainMenuKeyboard() *maxbot.Keyboard {
	kb := &maxbot.Keyboard{}
	kb.AddRow().AddCallback("📋 Создать задачу", schemes.POSITIVE, "create_task")
	kb.AddRow().AddCallback("📑 Мои задачи", schemes.DEFAULT, "my_tasks")
	return kb
}

func (h *Handler) taskCreatedKeyboard(taskID int) *maxbot.Keyboard {
	kb := &maxbot.Keyboard{}
	kb.AddRow().AddCallback("💬 Написать по задаче", schemes.DEFAULT, fmt.Sprintf("reply_task:%d", taskID))
	kb.AddRow().AddCallback("📂 Открыть", schemes.DEFAULT, fmt.Sprintf("open_task:%d", taskID))
	return kb
}

func (h *Handler) taskListFilterKeyboard() *maxbot.Keyboard {
	kb := &maxbot.Keyboard{}
	row := kb.AddRow()
	row.AddCallback("Актуальные", schemes.DEFAULT, "tasks_filter:active")
	row.AddCallback("Завершённые", schemes.DEFAULT, "tasks_filter:done")
	row.AddCallback("Все", schemes.DEFAULT, "tasks_filter:all")
	return kb
}

func (h *Handler) taskDetailKeyboard(taskID int) *maxbot.Keyboard {
	kb := &maxbot.Keyboard{}
	kb.AddRow().AddCallback("💬 Написать", schemes.DEFAULT, fmt.Sprintf("reply_task:%d", taskID))
	kb.AddRow().AddCallback("📎 Прикрепить файл", schemes.DEFAULT, fmt.Sprintf("attach_task:%d", taskID))
	return kb
}

func statusEmoji(s TaskStatus) string {
	switch s {
	case StatusOverdue:
		return "🔴"
	case StatusProgress:
		return "🟡"
	case StatusNew:
		return "🔵"
	case StatusReview:
		return "🟣"
	case StatusDone:
		return "🟢"
	default:
		return "⚪"
	}
}
