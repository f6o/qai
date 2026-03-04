package i18n

import (
	"sync"

	_ "embed"

	"unknwon.dev/i18n"
)

//go:embed locales/locale_en-US.ini
var localeData []byte

var (
	store *i18n.Store
	once  sync.Once
)

func Init() *i18n.Store {
	once.Do(func() {
		s := i18n.NewStore()

		// TODO: fall back to `en-US` for now
		_, err := s.AddLocale("en-US", "English", localeData)
		if err != nil {
			panic("failed to load locale: " + err.Error())
		}

		store = s
	})
	return store
}

func T(key string, args ...interface{}) string {
	if store == nil {
		Init()
	}
	l, err := store.Locale("en-US")
	if err != nil {
		return "messages::" + key
	}
	return l.Translate("messages::"+key, args...)
}
