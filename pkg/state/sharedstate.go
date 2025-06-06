// Copyright © 2024 JR team
//
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

package state

import (
	"math"
	"strings"
	"sync"
	"time"

	"github.com/biter777/countries"
	"github.com/jrnd-io/jrv2/pkg/random"
)

var _state *SharedState

type CSV map[string]string
type CSVMap map[int]CSV

type Execution struct {
	Start                     time.Time
	GeneratedObjects          uint64
	GeneratedBytes            uint64
	ExpectedObjects           int64
	CurrentIterationLoopIndex int
}

// SharedState is the object passed on the templates which contains all the needed details.
type SharedState struct {
	Execution    *Execution
	Locale       string
	Counters     sync.Map
	countersLock sync.RWMutex
	Ctx          sync.Map

	List     sync.Map
	listLock sync.RWMutex

	CSVMap  CSVMap
	csvLock sync.RWMutex

	LastIndex    int
	CountryIndex int
	CityIndex    int
}

func GetSharedState() *SharedState {

	cIndex := 0
	for i, c := range countries.All() {
		if c.Alpha2() == "US" {
			cIndex = i
		}
	}
	if _state == nil {
		_state = &SharedState{
			Execution: &Execution{
				Start: time.Now(),
			},
			Locale:       strings.ToLower(countries.UnitedStatesOfAmerica.Alpha2()),
			Counters:     sync.Map{},
			countersLock: sync.RWMutex{},
			Ctx:          sync.Map{},
			listLock:     sync.RWMutex{},
			List:         sync.Map{},
			CSVMap:       make(CSVMap),
			csvLock:      sync.RWMutex{},
			LastIndex:    -1,
			CountryIndex: cIndex, // int(countries.UnitedStatesOfAmerica),
			CityIndex:    -1,
		}
	}
	return _state
}

func (st *SharedState) AddValueToList(key string, value any) {
	st.listLock.Lock()
	defer st.listLock.Unlock()
	list, ok := st.List.Load(key)
	if !ok {
		list = []any{}
	}
	list = append(list.([]any), value)
	st.List.Store(key, list)
}

func (st *SharedState) RandomValueFromList(s string) any {
	st.listLock.Lock()
	defer st.listLock.Unlock()
	list, _ := st.List.Load(s)
	if list == nil {
		return ""
	}
	l := len(list.([]any))
	if l != 0 {
		return list.([]any)[random.Random.IntN(l)]
	}
	return ""
}

func (st *SharedState) GetValueFromListAtIndex(s string, index int) any {

	st.listLock.Lock()
	defer st.listLock.Unlock()
	list, _ := st.List.Load(s)
	if list == nil {
		return ""
	}
	l := len(list.([]any))
	if l != 0 && index < l {
		return list.([]any)[index]
	}

	return ""
}

// RandomNValuesFromList returns a random value from Context list l
func (st *SharedState) RandomNValuesFromList(s string, n int) []any {
	st.listLock.Lock()
	defer st.listLock.Unlock()
	list, _ := st.List.Load(s)
	if list == nil {
		return []any{""}

	}
	l := len(list.([]any))
	if l != 0 {
		ints := st.findNDifferentInts(n, l)
		results := make([]any, len(ints))
		for i := range ints {
			results[i] = list.([]any)[i]
		}
		return results
	}

	return []any{""}
}

// Helper function to check if an int is in a slice of ints
func contains(values []int, value int) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (st *SharedState) Counter(c string, start int, step int) int {
	st.countersLock.Lock()
	defer st.countersLock.Unlock()

	val, exists := st.Counters.Load(c)
	if exists {
		st.Counters.Store(c, val.(int)+step)
		return val.(int) + step
	}
	st.Counters.Store(c, start)
	return start
}

// Helper function to generate n different integers from 0 to length
func (st *SharedState) findNDifferentInts(n, max int) []int {

	n = int(math.Min(float64(n), float64(max)))
	ints := make([]int, n)

	// Generate n different random indices of maximum length
	for i := 0; i < n; {
		index := random.Random.IntN(max)
		if !contains(ints, index) {
			ints[i] = index
			i++
		}
	}

	return ints
}

func (st *SharedState) SetCSV(csv CSVMap) {
	st.CSVMap = csv

}

func (st *SharedState) Value(key string) any {
	value, _ := GetSharedState().Ctx.Load(key)
	return value
}

func (st *SharedState) FromCSV(c string) string {
	st.csvLock.Lock()
	defer st.csvLock.Unlock()

	if len(st.CSVMap) > 0 {
		return st.CSVMap[st.Execution.CurrentIterationLoopIndex%len(st.CSVMap)][c]
	}
	return ""

}
