/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package perfstat

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/op/go-logging"
)

const enableStats = false
const printPeriodically = true
const printInterval = 10000 //Millisecond
const commonPrefix = "github.com/cherryFloris/pbft-vrf/core/ledger"
const commonPrefixLen = len(commonPrefix)

var holder *statsHolder
var logger = logging.MustGetLogger("ledger.perfstat")

type statsHolder struct {
	rwLock sync.RWMutex
	m      map[string]*stat
}

func init() {
	if !enableStats {
		return
	}
	holder = &statsHolder{m: make(map[string]*stat)}
	if printPeriodically {
		go printStatsPeriodically()
	}
}

// UpdateTimeStat updates the stats for time spent at a particular point in the code
func UpdateTimeStat(id string, startTime time.Time) {
	updateStat(id, time.Since(startTime).Nanoseconds())
}

// UpdateDataStat updates the stats for data at a particular point in the code
func UpdateDataStat(id string, value int64) {
	updateStat(id, value)
}

// ResetStats resets all the stats data
func ResetStats() {
	if !enableStats {
		return
	}
	holder.rwLock.Lock()
	defer holder.rwLock.Unlock()
	for _, v := range holder.m {
		v.reset()
	}
}

// PrintStats prints the stats in the log file.
func PrintStats() {
	if !enableStats {
		return
	}
	holder.rwLock.RLock()
	defer holder.rwLock.RUnlock()
	logger.Info("Stats.......Start")
	var paths []string
	for k := range holder.m {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	for _, k := range paths {
		v := holder.m[k]
		logger.Info(v.String())
	}
	logger.Info("Stats.......Finish")
}

func updateStat(id string, value int64) {
	if !enableStats {
		return
	}
	path := getCallerInfo()
	statName := fmt.Sprintf("%s:%s", path, id)
	fmt.Println(statName)
	stat := getOrCreateStat(statName, "", 0)
	stat.updateDataStat(value)
}

func getOrCreateStat(name string, file string, line int) *stat {
	holder.rwLock.RLock()
	stat, ok := holder.m[name]
	if ok {
		holder.rwLock.RUnlock()
		return stat
	}

	holder.rwLock.RUnlock()
	holder.rwLock.Lock()
	defer holder.rwLock.Unlock()
	stat, ok = holder.m[name]
	if !ok {
		stat = newStat(name, fmt.Sprintf("%s:%d", file, line))
		holder.m[name] = stat
	}
	return stat
}

func printStatsPeriodically() {
	for {
		PrintStats()
		time.Sleep(time.Duration(int64(printInterval) * time.Millisecond.Nanoseconds()))
	}
}

func getCallerInfo() string {
	pc := make([]uintptr, 10)
	// Note: the default value 4 will ensure stat name exclude the path
	// "/perfstat.UpdateTimeStat -> /perfstat.updateStat -> /perfstat.getCallerInfo"
	// "/perfstat.UpdateDataStat -> /perfstat.updateStat -> /perfstat.getCallerInfo"
	runtime.Callers(4, pc)
	var path bytes.Buffer
	j := 0
	for i := range pc {
		f := runtime.FuncForPC(pc[i])
		funcName := f.Name()
		if strings.HasPrefix(funcName, commonPrefix) {
			j = i
		} else {
			break
		}
	}

	for i := j; i >= 0; i-- {
		f := runtime.FuncForPC(pc[i])
		funcName := f.Name()
		funcNameShort := funcName[commonPrefixLen:]
		path.WriteString(funcNameShort)
		if i > 0 {
			path.WriteString(" -> ")
		}
	}

	return path.String()
}
