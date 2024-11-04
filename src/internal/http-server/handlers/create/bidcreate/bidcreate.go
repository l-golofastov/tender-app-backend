package bidcreate

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
	"tender-app-backend/src/internal/storage"
)

type Request struct {
	Bid internal.Bid
}

type BidCreator interface {
	CreateBid(b internal.Bid) (internal.Bid, error)
}

func New(log *slog.Logger, bidCreator BidCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.create.bidcreate.New"

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

		// TODO: add correct validation
		if err := validator.New().Struct(req.Bid); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		bid, err := bidCreator.CreateBid(req.Bid)
		if err != nil {
			if errors.Is(err, storage.ErrOrgRespNotFound) {
				log.Info(
					"organisation responsible user not found",
					slog.String("creator_username", req.Bid.CreatorUsername),
				)

				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, response.Error("user is unable to create bids"))

				return
			}

			log.Error("failed to create bid", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to create bid"))

			return
		}

		log.Info("bid created", slog.Any("bid", bid))

		render.JSON(w, r, bid)
	}
}

