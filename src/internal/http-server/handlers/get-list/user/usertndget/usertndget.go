package usertndget

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
)

type UserTenderGetter interface {
	GetUserTendersList(username string) ([]internal.Tender, error)
}

func New(log *slog.Logger, tenderGetter UserTenderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.user.usertndget.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		username := r.URL.Query().Get("username")
		if username == "" {
			log.Info("username is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		res, err := tenderGetter.GetUserTendersList(username)
		if err != nil {
			log.Error("failed to get user tenders list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get user tenders list"))

			return
		}

		render.JSON(w, r, res)
	}
}
