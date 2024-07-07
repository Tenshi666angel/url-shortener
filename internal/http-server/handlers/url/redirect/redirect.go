package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"shortener/internal/constants"
	resp "shortener/internal/lib/api/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UrlGetter interface {
	GetUrl(alias string) (string, error)
}

func New(logger *slog.Logger, urlGetter UrlGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			logger.Info("alias is empty")

			render.JSON(w, r, resp.Error("bad request"))

			return
		}

		resUrl, err := urlGetter.GetUrl(alias)

		if errors.Is(err, constants.ErrUrlNotFound) {
			logger.Info("url not found", slog.String("alias", alias))

			render.JSON(w, r, resp.Error("url not found"))

			return
		}

		if err != nil {
			logger.Error("failed to get url")

			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		logger.Info("got url", slog.String("url", resUrl))

		http.Redirect(w, r, resUrl, http.StatusFound)
	}
}
