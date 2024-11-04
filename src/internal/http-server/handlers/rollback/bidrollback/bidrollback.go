package bidrollback

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
	"tender-app-backend/src/internal/storage"
)

type BidRollbacker interface {
	RollbackBid(bidId, version int) (internal.Bid, error)
}

func New(log *slog.Logger, bidRollbacker BidRollbacker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.rollback.bidrollback.New"

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

		versionStr := chi.URLParam(r, "version")
		if versionStr == "" {
			log.Info("version is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Info("failed to parse version")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		bid, err := bidRollbacker.RollbackBid(bidId, version)
		if err != nil {
			if errors.Is(err, storage.ErrBidNotFound) {
				log.Info(
					"bid not found",
					slog.String("bid_id", bidIdStr),
					slog.String("version", versionStr),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("bid or its version not found"))

				return
			}

			log.Error("failed to roll back bid", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to roll back bid"))

			return
		}

		log.Info("bid rolled back", slog.Any("bid", bid))

		render.JSON(w, r, bid)
	}
}
