package bidstatus

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"tender-app-backend/src/internal"
	"tender-app-backend/src/internal/lib/api/response"
	"tender-app-backend/src/internal/lib/logger/sl"
)

type BidStatusGetter interface {
	GetBidsList() ([]internal.Bid, error)
}

type Response struct {
	Id     int
	Status string
}

func New(log *slog.Logger, bidGetter BidStatusGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get-list.status.bidstatus.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		resp := make([]Response, 0)

		res, err := bidGetter.GetBidsList()
		if err != nil {
			log.Error("failed to get bids status list", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get bids status list"))

			return
		}

		for _, b := range res {
			r := Response{
				Id:     b.Id,
				Status: b.Status,
			}
			resp = append(resp, r)
		}

		render.JSON(w, r, resp)
	}
}
