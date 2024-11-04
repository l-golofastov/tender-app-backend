package tndstatus

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
)

type TenderStatusGetter interface {
	GetTendersList() ([]internal.Tender, error)
}

type Response struct {
	Id     int
	Status string
}

func New(log *slog.Logger, tenderGetter TenderStatusGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.status.tndstatus.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		resp := make([]Response, 0)

		res, err := tenderGetter.GetTendersList()
		if err != nil {
			log.Error("failed to get tenders status list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get tenders status list"))

			return
		}

		for _, t := range res {
			r := Response{
				Id:     t.Id,
				Status: t.Status,
			}
			resp = append(resp, r)
		}

		render.JSON(w, r, resp)
	}
}
