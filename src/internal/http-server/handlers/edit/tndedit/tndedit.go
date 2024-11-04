package tndedit

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
	Tender internal.Tender
}

type TenderEditor interface {
	EditTender(t internal.Tender, editId int) (internal.Tender, error)
}

func New(log *slog.Logger, tenderEditor TenderEditor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.edit.tndedit.New"

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

		tender, err := tenderEditor.EditTender(req.Tender, tenderId)
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

			log.Error("failed to edit tender", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to edit tender"))

			return
		}

		log.Info("tender edited", slog.Any("tender", tender))

		render.JSON(w, r, tender)
	}
}
