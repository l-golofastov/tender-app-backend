package bidget

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

type BidGetter interface {
	GetTenderBidsList(tenderId int) ([]internal.Bid, error)
}

func New(log *slog.Logger, bidGetter BidGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.all.bidget.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		tenderIdStr := chi.URLParam(r, "tenderId")
		if tenderIdStr == "" {
			log.Info("tender id is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		tenderId, err := strconv.Atoi(tenderIdStr)
		if err != nil {
			log.Info("failed to parse tender id")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		res, err := bidGetter.GetTenderBidsList(tenderId)
		if err != nil {
			if errors.Is(err, storage.ErrTenderNotFound) {
				log.Info(
					"tender not found",
					slog.String("tender_id", tenderIdStr),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("tender not found"))

				return
			}

			log.Error("failed to get tenders list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get tenders list"))

			return
		}

		render.JSON(w, r, res)
	}
}
