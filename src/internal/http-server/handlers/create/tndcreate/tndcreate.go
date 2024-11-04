package tndcreate

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
	Tender internal.Tender
}

type TenderCreator interface {
	CreateTender(t internal.Tender) (internal.Tender, error)
}

func New(log *slog.Logger, tenderCreator TenderCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.create.tndcreate.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req.Tender)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// TODO: add correct validation
		if err := validator.New().Struct(req.Tender); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		tender, err := tenderCreator.CreateTender(req.Tender)
		if err != nil {
			if errors.Is(err, storage.ErrOrgRespNotFound) {
				log.Info(
					"organisation responsible user not found",
					slog.String("creator_username", req.Tender.CreatorUsername),
				)

				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, response.Error("user is unable to create tenders"))

				return
			}

			log.Error("failed to create tender", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to create tender"))

			return
		}

		log.Info("tender created", slog.Any("tender", tender))

		render.JSON(w, r, tender)
	}
}
