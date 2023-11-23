package signal

import (
	"os"
	"os/signal"
)

type HandleFunc func()

type Handler struct {
	regSigMap  map[os.Signal]HandleFunc
	listenChan chan os.Signal
}

func (t *Handler) InitSignalHandler(listenChanSize uint32) error {
	t.regSigMap = make(map[os.Signal]HandleFunc)
	t.listenChan = make(chan os.Signal, listenChanSize)
	return nil
}

func (t *Handler) RegisterSignal(signal os.Signal, handleFunc HandleFunc) bool {
	if _, exist := t.regSigMap[signal]; exist {
		return false
	}

	t.regSigMap[signal] = handleFunc
	return true
}

func (t *Handler) CloseSignalHandler() {
	if t.listenChan == nil {
		return
	}
	close(t.listenChan)
}

func (t *Handler) ListenSignal() {
	var sigs []os.Signal
	for sig := range t.regSigMap {
		sigs = append(sigs, sig)
	}
	signal.Notify(t.listenChan, sigs...)

	for {
		sig, ok := <-t.listenChan
		if !ok {
			break
		}

		handleFunc, exist := t.regSigMap[sig]
		if !exist {
			continue
		}

		handleFunc()
	}

}
