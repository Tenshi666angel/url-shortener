package save

import (
	"errors"
	"log/slog"
	"net/http"

	"shortener/internal/constants"
	resp "shortener/internal/lib/api/response"
	"shortener/internal/lib/logger/sl"
	"shortener/internal/lib/random"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const aliasLength = 10

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type UrlSaver interface {
	SaveUrl(url string, alias string) (int64, error)
}

func New(logger *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			logger.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			logger.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveUrl(req.Url, alias)

		if errors.Is(err, constants.ErrUrlExists) {
			logger.Info("url already exists", slog.String("url", req.Url))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		if err != nil {
			logger.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		_ = id

		render.JSON(w, r, Response{
			Response: resp.Ok(),
			Alias:    alias,
		})
	}
}
