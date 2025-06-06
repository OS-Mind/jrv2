package loop

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/jrnd-io/jrv2/pkg/state"

	"github.com/jrnd-io/jrv2/pkg/emitter"
	"github.com/jrnd-io/jrv2/pkg/plugin"
	"github.com/rs/zerolog/log"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func DoLoop(ctx context.Context,
	emitters *orderedmap.OrderedMap[string, []emitter.Config],
	configParams map[string]string,
	pluginName string,
	pluginLogLevel hclog.Level) error {

	// emitter slice
	es := make([]*emitter.Emitter, 0)

	//  ctrl-c signal
	controlC, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// wait group to synchronize tickers end
	var wg sync.WaitGroup

	pluginMap := make(map[string]*plugin.Plugin)
	defer func() {
		for _, _p := range pluginMap {
			err := _p.Close()
			if err != nil {
				log.Warn().Err(err).Str("plugin", _p.Name).Msg("error in closing plugin")
			}
		}
	}()

	// starting loop
	for e := emitters.Oldest(); e != nil; e = e.Next() {
		log.Debug().
			Str("emitter", e.Key).
			Int("len", len(e.Value)).
			Msg("Starting loop for emitters")

		for i, cfg := range e.Value {

			log.Debug().
				Int("emitter", i).
				Interface("config", cfg).
				Msg("Running emitter")
			em, err := emitter.NewFromConfig(cfg)
			if err != nil {
				return err
			}

			es = append(es, em) //nolint
			// choosing output either from emitter or from passed value
			output := em.Config.Output
			if pluginName != "" {
				output = pluginName
			}

			// setting plugin or get it from map
			var _plugin *plugin.Plugin
			if pluginMap[output] == nil {
				log.Debug().
					Str("output", output).
					Msg("creating emitter output")
				_plugin, err = plugin.New(output, pluginLogLevel)
				if err != nil {
					return err
				}

				pluginMap[output] = _plugin
			} else {
				log.Debug().
					Str("output", output).
					Msg("reusing emitter output")
				_plugin = pluginMap[output]
			}
			log.Debug().
				Str("emitter", em.Config.Name).
				Str("plugin", _plugin.Name).Msg("setting emitter plugin")
			em.SetPlugin(_plugin)

			wg.Add(1)
			go func(e *emitter.Emitter) {
				defer wg.Done()

				frequency := e.Config.Tick.Frequency
				if frequency > 0 {
					log.Debug().
						Dur("frequency", frequency).
						Str("emitter", e.Config.Name).
						Msg("Starting ticker")
					//					ticker := time.NewTicker(frequency)
					//					defer ticker.Stop()
					e.StartTicker()

					for {
						select {
						case <-controlC.Done():
							stop()
							return
						case <-e.Ticker.C:
							doTemplate(ctx, e, configParams)
						case <-e.StopChannel:
							return
						}

					}
				} else {
					log.Debug().
						Str("Emitter: %e", e.Config.Name).
						Msg("Exec do Template")
					doTemplate(ctx, e, configParams)
				}
			}(es[i])

		}
	}

	wg.Wait()
	return nil
}

func doTemplate(ctx context.Context, em *emitter.Emitter, configParams map[string]string) { //nolint

	var err error

	localState := state.NewState()
	for i := 0; i < em.Config.Tick.Num; i++ {
		state.GetSharedState().Execution.CurrentIterationLoopIndex++

		keyText := ""
		valueText := ""

		if em.ValueTemplate != nil {
			valueText = em.ValueTemplate.ExecuteWith(localState)
			if em.Config.Oneline {
				valueText = strings.ReplaceAll(valueText, "\n", "")
			}
		}
		if em.KeyTemplate != nil {
			keyText = em.KeyTemplate.Execute()
			log.Debug().Str("key", keyText).Msg("key generated with template")
		} else {
			keyText = localState.Key
			log.Debug().Str("key", keyText).Msg("key generated within localState")
		}

		// building emitter configuration map
		cfgParams := make(map[string]string)
		for k, v := range em.Config.ConfigParameters {
			cfgParams[k] = v
		}
		cfgParams["emitter.name"] = em.Config.Name
		for k, v := range configParams {
			ks := strings.Split(k, ".")
			if len(ks) == 1 {
				cfgParams[k] = v
			} else if ks[0] == em.Config.Name {
				log.Debug().
					Str("key", ks[1]).
					Str("value", v).
					Str("name", em.Config.Name).
					Msg("adding configuration parameter")
				wholeKey := strings.Join(ks[1:], ".")
				cfgParams[wholeKey] = v
			}
		}

		log.Debug().
			Str("name", em.Config.Name).
			Interface("cfgParams", cfgParams).Msg("configuration parameters")

		var resp *jrpc.ProduceResponse
		resp, err = em.Produce(ctx, []byte(keyText), []byte(valueText), localState.Header, cfgParams)
		if err != nil {
			log.Warn().
				Err(err).
				Str("name", em.Config.Name).
				Msg("error in emission")
		} else {
			state.GetSharedState().Execution.GeneratedObjects++
			state.GetSharedState().Execution.GeneratedBytes += resp.Bytes
		}
	}

}
