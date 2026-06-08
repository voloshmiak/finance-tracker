package controller

import (
	"errors"
	"finance-tracker/src/usecase"
	"net/http"
)

type ReminderController struct {
	useCase *usecase.Reminder
}

func NewReminderController(useCase *usecase.Reminder) *ReminderController {
	return &ReminderController{
		useCase: useCase,
	}
}

// ServeHTTP godoc
// @Summary     Send Reminder
// @Description Triggers a Telegram reminder about card reports not yet uploaded this month. Intended to be called by a scheduled job.
// @Tags        Cron
// @Param       Authorization  header  string  true  "Bearer <auth token>"
// @Success     200            "Reminder sent"
// @Failure     401   {object}  string  "Unauthorized"
// @Failure     500   {object}  string  "Internal Server Error"
// @Router      /cron/reminder [post]
func (c *ReminderController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := c.useCase.Execute(r.Context(), r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, usecase.ErrUnauthorized) {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (c *ReminderController) Register(mux *http.ServeMux) {
	mux.Handle("POST /cron/reminder", c)
}
