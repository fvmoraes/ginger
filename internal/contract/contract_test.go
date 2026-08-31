// Package contract pin the public API surface of pkg/ (Fase 0, tarefa 5 —
// inventário de APIs públicas e teste de contrato mínimo).
//
// O teste abaixo referencia cada símbolo exportado que constitui compromisso
// de compatibilidade. Qualquer remoção/renomeação de API pública quebra a
// compilação deste teste — a mudança precisa ser consciente (deprecação
// documentada em vez de quebra silenciosa).
package contract

import (
	"github.com/fvmoraes/ginger/pkg/app"
	"github.com/fvmoraes/ginger/pkg/config"
	"github.com/fvmoraes/ginger/pkg/database"
	"github.com/fvmoraes/ginger/pkg/errors"
	"github.com/fvmoraes/ginger/pkg/health"
	"github.com/fvmoraes/ginger/pkg/logger"
	"github.com/fvmoraes/ginger/pkg/middleware"
	"github.com/fvmoraes/ginger/pkg/response"
	"github.com/fvmoraes/ginger/pkg/router"
	"github.com/fvmoraes/ginger/pkg/sse"
	"github.com/fvmoraes/ginger/pkg/testhelper"
	"github.com/fvmoraes/ginger/pkg/ws"
	"testing"
)

// pkg/app
var (
	_ = app.New
	_ = (*app.App).Run
	_ = (*app.App).OnStop
)

// pkg/config
var (
	_ = config.Load
	_ config.Config
	_ config.AppConfig
	_ config.HTTPConfig
	_ config.DatabaseConfig
	_ config.LogConfig
)

// pkg/database
var (
	_ = database.Connect
	_ = database.NewChecker
	_ = (*database.Checker).Check
	_ = (*database.Checker).Name
	_ database.Config
)

// pkg/errors
var (
	_ = errors.BadRequest
	_ = errors.Conflict
	_ = errors.Forbidden
	_ = errors.As
	_ = (*errors.AppError).Error
	_ = (*errors.AppError).Is
	_ = (*errors.AppError).Unwrap
	_ = (*errors.AppError).HTTPStatus
)

// pkg/health
var (
	_ = health.New
	_ = (*health.Handler).Register
	_ = (*health.Handler).ServeHTTP
	_ health.Checker
	_ health.Response
	_ health.Status
)

// pkg/logger
var (
	_ = logger.New
	_ = logger.FromContext
	_ = (*logger.Logger).With
)

// pkg/middleware
var (
	_ = middleware.Chain
	_ = middleware.Logger
	_ = middleware.CORS
	_ = middleware.Recover
	_ = middleware.RequestID
	_ = middleware.RequestIDFromContext
	_ middleware.CORSConfig
)

// pkg/response
var (
	_ = response.OK[any]
	_ = response.Created[any]
	_ = response.NoContent
	_ = response.Paginated[any]
	_ response.Envelope[any]
	_ response.Meta
	_ response.Page[any]
)

// pkg/router
var (
	_ = router.New
	_ = (*router.Router).GET
	_ = (*router.Router).Group
	_ = (*router.Router).Handle
	_ = router.Decode
	_ = router.JSON
	_ = router.Error
)

// pkg/sse
var (
	_ = sse.New
	_ = (*sse.Stream).Send
	_ = (*sse.Stream).SendComment
	_ sse.Event
	_ sse.Stream
)

// pkg/ws
var (
	_ = ws.Handle
	_ = (*ws.Conn).Send
	_ = (*ws.Conn).Recv
	_ = (*ws.Conn).Close
	_ ws.Handler
	_ ws.Message
)

// pkg/testhelper
var (
	_ = testhelper.NewRequest
	_ = (*testhelper.Request).Do
	_ = (*testhelper.Request).WithBody
	_ = (*testhelper.Request).WithHeader
	_ = testhelper.AssertStatus
	_ = testhelper.AssertHeader
	_ = testhelper.DecodeJSON
)

// TestContractCompiles existe para que o pacote tenha um teste executável;
// o contrato real é a própria compilação das referências acima.
func TestContractCompiles(t *testing.T) {
	t.Parallel()
}
