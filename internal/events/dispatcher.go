package events

import (
  "strconv"
  "sync"
)

// ------------- 全局处理器
var (
  hMu      sync.RWMutex
  handlers []Handler
)

// Register adds a new Handler (global, for logging/audit etc).
func Register(h Handler) {
  hMu.Lock()
  handlers = append(handlers, h)
  hMu.Unlock()
}

// ------------- 基于 name+type 的一次性订阅 -------------

// key: name + "#" + type
func key(name string, t EventType) string {
  return name + "#" + string(t)
}

var (
  sMu  sync.Mutex
  subs = make(map[string][]chan Event) // per key subscribers (once)
)

// SubscribeOnce subscribes to a single (name,type) event.
// The returned channel will receive *one* matching event then be closed.
// 典型用法：等待某个进程的 process.started，再把结果回写给客户端。
func SubscribeOnce(name string, t EventType) <-chan Event {
  ch := make(chan Event, 1)
  k := key(name, t)
  sMu.Lock()
  subs[k] = append(subs[k], ch)
  sMu.Unlock()
  return ch
}

// ------------- Emit 分发（同时支持全局 Handler 与一次性订阅） -------------

// Emit dispatches the Event to all registered handlers and notifies subscribers.
func Emit(e Event) {
  // 1) 复制 handlers 快照，避免持锁执行用户代码
  hMu.RLock()
  hcopy := make([]Handler, len(handlers))
  copy(hcopy, handlers)
  hMu.RUnlock()

  // 2) 触发一次性订阅者（只触发一次就移除）
  k := key(e.Name, e.Type)
  var toNotify []chan Event
  sMu.Lock()
  if list, ok := subs[k]; ok && len(list) > 0 {
    toNotify = list
    delete(subs, k) // once 语义：本次触发后清掉
  }
  sMu.Unlock()

  // 通知订阅者（不在持锁状态下）
  for _, ch := range toNotify {
    ch <- e
    close(ch)
  }

  // 3) 通知全局 handlers（为避免阻塞，每个 handler 起 goroutine，保持你原先语义）
  for _, h := range hcopy {
    h := h
    go h.Handle(e)
  }
}

// Stats 返回当前已注册的全局 handler 数量与订阅数（用于调试）
func Stats() (handlersN int, subsN int) {
  hMu.RLock()
  handlersN = len(handlers)
  hMu.RUnlock()

  sMu.Lock()
  for _, list := range subs {
    subsN += len(list)
  }
  sMu.Unlock()
  return
}

// DumpKeys 列出当前订阅键（仅用于调试）
func DumpKeys() []string {
  sMu.Lock()
  keys := make([]string, 0, len(subs))
  for k, list := range subs {
    keys = append(keys, k+"("+strconv.Itoa(len(list))+")")
  }
  sMu.Unlock()
  return keys
}
