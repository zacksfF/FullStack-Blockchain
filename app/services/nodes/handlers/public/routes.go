package public

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zacksfF/FullStack-Blockchain/blockchain/state"
	"github.com/zacksfF/FullStack-Blockchain/events"
	"github.com/zacksfF/FullStack-Blockchain/nameservices"
	"github.com/zacksfF/FullStack-Blockchain/web2"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservices.NameService
	Evts  *events.Events
}

// Routes binds all the public routes.
func Routes(app *web2.App, cfg Config) {
	pbl := Handlers{
		Log:   cfg.Log,
		State: cfg.State,
		NS:    cfg.NS,
		WS:    websocket.Upgrader{},
		Evts:  cfg.Evts,
	}

	const version = "v1"

	app.Handle(http.MethodGet, version, "/events", pbl.Events)
	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)
	app.Handle(http.MethodGet, version, "/accounts/list", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/accounts/list/:account", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/blocks/list", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/blocks/list/:account", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", pbl.Mempool)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list/:account", pbl.Mempool)
	app.Handle(http.MethodPost, version, "/tx/submit", pbl.SubmitWalletTransaction)
	app.Handle(http.MethodPost, version, "/tx/proof/:block/", pbl.SubmitWalletTransaction)
}
