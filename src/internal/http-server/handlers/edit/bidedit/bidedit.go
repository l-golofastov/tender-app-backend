package bidedit

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

type Request struct {
	Bid internal.Bid
}

type BidEditor interface {
	EditBid(b internal.Bid, editId int) (internal.Bid, error)
}

func New(log *slog.Logger, bidEditor BidEditor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.edit.bidedit.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req.Bid)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

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

		bid, err := bidEditor.EditBid(req.Bid, bidId	)
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

			log.Error("failed to edit bid", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to edit bid"))

			return
		}

		log.Info("bid edited", slog.Any("bid", bid))

		render.JSON(w, r, bid)
	}
}

