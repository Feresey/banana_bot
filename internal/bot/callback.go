package bot

import (
	"sync"
	"time"

	tgbotapi "github.com/Feresey/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type callbackStorage struct {
	log *zap.Logger

	mu        sync.RWMutex
	callbacks map[int]callbackFunc
}

func NewCallbackStorage(log *zap.Logger) *callbackStorage {
	return &callbackStorage{
		log:       log.Named("callback"),
		callbacks: make(map[int]callbackFunc),
	}
}

type callbackFunc func(*tgbotapi.CallbackQuery)

func (cs *callbackStorage) AddCallback(messageID int, cb callbackFunc, expire time.Duration) {
	cs.mu.Lock()
	cs.callbacks[messageID] = cb
	cs.mu.Unlock()
	time.AfterFunc(expire, func() {
		cs.mu.Lock()
		delete(cs.callbacks, messageID)
		cs.mu.Unlock()
	})
}

func (cs *callbackStorage) HandleCallabck(cbQuery *tgbotapi.CallbackQuery) {
	if cbQuery.Message == nil {
		return
	}
	cs.mu.RLock()
	cb, ok := cs.callbacks[cbQuery.Message.MessageID]
	cs.mu.RUnlock()
	if ok && cb != nil {
		go cb(cbQuery)
	}
}
