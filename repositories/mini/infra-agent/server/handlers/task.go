package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"infra-agent/server/tasks"
)

// TaskHandler 处理任务相关的 HTTP 请求。
type TaskHandler struct {
	taskService tasks.TaskService
}

// NewTaskHandler 创建任务处理器。
func NewTaskHandler(taskService tasks.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create 创建一个任务。
func (h *TaskHandler) Create(c *gin.Context) {
	task, err := h.taskService.Create(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "create task"})
		return
	}
	c.JSON(http.StatusCreated, task)
}
