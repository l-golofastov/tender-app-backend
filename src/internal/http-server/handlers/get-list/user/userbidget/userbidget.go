package userbidget

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
)

type UserBidGetter interface {
	GetUserBidsList(username string) ([]internal.Bid, error)
}

func New(log *slog.Logger, bidGetter UserBidGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.user.userbidget.New"

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

		res, err := bidGetter.GetUserBidsList(username)
		if err != nil {
			log.Error("failed to get user bids list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get user bids list"))

			return
		}

		render.JSON(w, r, res)
	}
}
