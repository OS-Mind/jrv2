// Copyright © 2024 JR team
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package luascript

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/jrnd-io/jrv2/pkg/jrpc"
	"github.com/jrnd-io/jrv2/pkg/plugin"
	"github.com/rs/zerolog/log"
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	luar "layeh.com/gopher-luar"
)

const (
	ScriptName = "luajr"
	PluginName = "luascript"
)

func init() {
	p, err := New()
	if err == nil {
		plugin.RegisterLocalPlugin(PluginName, p)
	}
}

type Plugin struct {
	configuration    Config
	luaProtoFunction *lua.FunctionProto
}

func New() (*Plugin, error) {

	cfgBytes, err := plugin.ReadConfig(PluginName)
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config")
	}

	p := &Plugin{
		configuration: config,
	}
	p.InitializeFromConfig(config)
	return p, nil
}

func (p *Plugin) InitializeFromConfig(config Config) {

	var err error
	if config.Script == "" && config.ScriptFile == "" {
		log.Fatal().Msg("script or script_file is required")
	}

	var scriptBytes []byte
	if config.ScriptFile != "" {
		scriptBytes, err = os.ReadFile(config.ScriptFile)
		if err != nil {
			log.Fatal().Err(err).Str("script_file", config.ScriptFile).Msg("Failed to read script file")
		}
	} else {
		scriptBytes = []byte(config.Script)
	}

	chunk, err := parse.Parse(strings.NewReader(string(scriptBytes)), ScriptName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse script")
	}
	p.luaProtoFunction, err = lua.Compile(chunk, ScriptName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to compile script")
	}

	p.configuration = config

}

func (p *Plugin) Produce(_ context.Context, k []byte, v []byte, h map[string]string) (*jrpc.ProduceResponse, error) {

	L := lua.NewState()
	libs.Preload(L)

	L.SetGlobal("k", lua.LString(k))
	L.SetGlobal("v", lua.LString(string(v)))
	L.SetGlobal("headers", luar.New(L, h))

	lf := L.NewFunctionFromProto(p.luaProtoFunction)
	L.Push(lf)
	err := L.PCall(0, 0, nil)
	if err != nil {
		return nil, err
	}
	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: "",
	}, nil

}
