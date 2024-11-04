package submit

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
	"tender-app-backend/src/internal/storage"
)

type Submitter interface {
	SubmitBid(bidId int, orgUsername string) error
}

func New(log *slog.Logger, submitter Submitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.submit.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		bidIdStr := chi.URLParam(r, "bidId")
		if bidIdStr == "" {
			log.Info("bid id is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		bidId, err := strconv.Atoi(bidIdStr)
		if err != nil {
			log.Info("failed to parse bid id")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			log.Info("username is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		err = submitter.SubmitBid(bidId, username)
		if err != nil {
			if errors.Is(err, storage.ErrBidNotFound) {
				log.Info(
					"bid not found",
					slog.String("bid_id", bidIdStr),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("bid not found"))

				return
			}

			if errors.Is(err, storage.ErrBidNotPublished) {
				log.Info(
					"bid not published",
					slog.String("bid_id", bidIdStr),
				)

				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, response.Error("bid not published"))

				return
			}

			if errors.Is(err, storage.ErrTenderNotPublished) {
				log.Info(
					"bid tender not published",
					slog.String("bid_id", bidIdStr),
				)

				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, response.Error("bid tender not published"))

				return
			}

			if errors.Is(err, storage.ErrOrgRespNotFound) {
				log.Info(
					"organisation responsible user not found",
					slog.String("username", username),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("organisation responsible user not found"))

				return
			}

			log.Error("failed to submit bid", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to submit bid"))

			return
		}

		log.Info("bid submitted", slog.String("bidId", bidIdStr))

		render.JSON(w, r, response.OK())
	}
}
