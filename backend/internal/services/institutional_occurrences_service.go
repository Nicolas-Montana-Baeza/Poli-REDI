package services

import (
	"errors"
	"strings"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

type InstitutionalOccurrenceRequest struct {
	ScheduleID   int     `json:"scheduleId"`
	OriginalDate string  `json:"originalDate"`
	Action       string  `json:"action"`
	Reason       string  `json:"reason"`
	NewDate      *string `json:"newDate,omitempty"`
	NewStartTime *string `json:"newStartTime,omitempty"`
	NewEndTime   *string `json:"newEndTime,omitempty"`
}

func validateOccurrenceRequest(req *InstitutionalOccurrenceRequest) error {
	req.Action = strings.ToUpper(strings.TrimSpace(req.Action))
	req.OriginalDate = strings.TrimSpace(req.OriginalDate)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.ScheduleID <= 0 || len(req.Reason) < 5 || len(req.Reason) > 500 {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	if _, err := time.Parse("2006-01-02", req.OriginalDate); err != nil {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	if req.Action == "CANCEL" {
		if req.NewDate != nil || req.NewStartTime != nil || req.NewEndTime != nil {
			return repositories.ErrInstitutionalOccurrenceInvalid
		}
		return nil
	}
	if req.Action != "RESCHEDULE" || req.NewDate == nil || req.NewStartTime == nil || req.NewEndTime == nil {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	*req.NewDate = strings.TrimSpace(*req.NewDate)
	*req.NewStartTime = strings.TrimSpace(*req.NewStartTime)
	*req.NewEndTime = strings.TrimSpace(*req.NewEndTime)
	if _, err := time.Parse("2006-01-02", *req.NewDate); err != nil {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	start, err := time.Parse("15:04", *req.NewStartTime)
	if err != nil {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	end, err := time.Parse("15:04", *req.NewEndTime)
	if err != nil || !end.After(start) {
		return repositories.ErrInstitutionalOccurrenceInvalid
	}
	return nil
}

func GetInstitutionalOccurrenceExceptions(user models.LocalAuthUser, activityID int) ([]repositories.InstitutionalOccurrenceException, error) {
	if _, err := institutionalActivityForManagement(user, activityID); err != nil {
		return nil, err
	}
	return repositories.GetInstitutionalOccurrenceExceptions(activityID)
}

func PatchInstitutionalOccurrence(user models.LocalAuthUser, activityID int, req InstitutionalOccurrenceRequest) (repositories.InstitutionalOccurrenceException, error) {
	activity, err := institutionalActivityForManagement(user, activityID)
	if err != nil {
		return repositories.InstitutionalOccurrenceException{}, err
	}
	if activity.Status != models.InstitutionalActivityStatusScheduled {
		return repositories.InstitutionalOccurrenceException{}, repositories.ErrInstitutionalOccurrenceInvalid
	}
	if err = validateOccurrenceRequest(&req); err != nil {
		return repositories.InstitutionalOccurrenceException{}, err
	}
	if req.Action == "RESCHEDULE" {
		schedule := models.CreateInstitutionalScheduleRequest{
			StartTime: *req.NewStartTime, EndTime: *req.NewEndTime,
		}
		if err := ensureInstitutionalOperatingHours([]models.CreateInstitutionalScheduleRequest{schedule}); err != nil {
			return repositories.InstitutionalOccurrenceException{}, err
		}
		endClock, _ := time.Parse("15:04", *req.NewEndTime)
		date, _ := time.ParseInLocation("2006-01-02", *req.NewDate, businessclock.Location())
		end := time.Date(date.Year(), date.Month(), date.Day(), endClock.Hour(), endClock.Minute(), 0, 0, businessclock.Location())
		if !end.After(businessclock.Now()) {
			return repositories.InstitutionalOccurrenceException{}, repositories.ErrInstitutionalOccurrenceInvalid
		}
	}
	result, err := repositories.ApplyInstitutionalOccurrenceException(activityID, user.ID,
		repositories.InstitutionalOccurrencePatch{
			ScheduleID: req.ScheduleID, OriginalDate: req.OriginalDate, Action: req.Action,
			Reason: req.Reason, NewDate: req.NewDate, NewStartTime: req.NewStartTime, NewEndTime: req.NewEndTime,
		})
	if errors.Is(err, repositories.ErrInstitutionalOccurrenceInvalid) {
		return result, err
	}
	return result, err
}
