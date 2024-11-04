package tndrollback

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

type TenderRollbacker interface {
	RollbackTender(tenderId, version int) (internal.Tender, error)
}

func New(log *slog.Logger, tenderRollbacker TenderRollbacker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.rollback.tndrollback.New"

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

		tender, err := tenderRollbacker.RollbackTender(tenderId, version)
		if err != nil {
			if errors.Is(err, storage.ErrTenderNotFound) {
				log.Info(
					"tender not found",
					slog.String("tender_id", tenderIdStr),
					slog.String("version", versionStr),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("tender or its version not found"))

				return
			}

			log.Error("failed to roll back tender", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to roll back tender"))

			return
		}

		log.Info("tender rolled back", slog.Any("tender", tender))

		render.JSON(w, r, tender)
	}
}

