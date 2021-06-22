package udb

import (
	"container/heap"
	_ "fmt"
	"log"
	_ "runtime/debug"
	_ "sync"
	"time"
)

const MIN_TIMER = 20 * time.Millisecond

var (
	logger    *log.Logger
	IsRunning bool = true
)

func (h *UserData) GlobalTick() {
	now := time.Now()
	for {
		if h.Len() <= 0 {
			break
		}
		nextRunTime := h.timers[0].runTime
		if nextRunTime.After(now) {
			// not due time
			break
		}
		t := heap.Pop(h).(*TimerEntry)

		if t.taskId != "" {
			delete(h.TaskTimerEntryMap, t.taskId)
		}

		callback := t.callback
		if callback == nil {
			continue
		}

		callback()
	}
}

func (u *UserData) Print() {
	//fmt.Println("user id : ", u.UserData().(*context.AgentInfo).UID)
}

func (h *UserData) Len() int {
	return len(h.timers)
}

func (h *UserData) Less(i, j int) bool {
	t1, t2 := h.timers[i].runTime, h.timers[j].runTime
	if t1.Before(t2) {
		return true
	}
	return false
}

func (h *UserData) Swap(i, j int) {
	var tmp *TimerEntry
	tmp = h.timers[i]
	h.timers[i] = h.timers[j]
	h.timers[j] = tmp
}

func (h *UserData) Push(x interface{}) {
	h.timers = append(h.timers, x.(*TimerEntry))
}

func (h *UserData) Pop() (ret interface{}) {
	l := len(h.timers)
	h.timers, ret = h.timers[:l-1], h.timers[l-1]
	return
}

func (h *UserData) GetLength() map[string]int {
	return map[string]int{
		"heap_len": h.Len(),
	}
}

func (h *UserData) RemoveById(id string) {
	if entry, ok := h.TaskTimerEntryMap[id]; ok {
		entry.runTime = time.Now() //将time 设置成now， 排序立刻会被放到最前或者最后
		entry.callback = nil
	}
}

func (h *UserData) addCallback(d time.Duration, taskId string, callback CallbackFunc) *TimerEntry {
	if d < MIN_TIMER {
		d = MIN_TIMER
	}

	t := &TimerEntry{
		runTime:  time.Now().Add(d),
		taskId:   taskId,
		callback: callback,
	}

	h.TaskTimerEntryMap[taskId] = t
	heap.Push(h, t)
	return t
}

func (h *UserData) AddTimer(dueInterval int, taskId string, loop bool, callback CallbackFunc) {
	h.addCallback(time.Duration(dueInterval)*time.Second, taskId, func() {
		callback()
		if loop && h != nil {
			h.AddTimer(dueInterval, taskId, loop, callback)
		}
	})
}

type CallbackFunc func()
