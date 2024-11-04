package tndget

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
)

type TenderGetter interface {
	GetTendersList() ([]internal.Tender, error)
}

func New(log *slog.Logger, tenderGetter TenderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.all.tndget.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		res, err := tenderGetter.GetTendersList()
		if err != nil {
			log.Error("failed to get tenders list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get tenders list"))

			return
		}

		render.JSON(w, r, res)
	}
}
