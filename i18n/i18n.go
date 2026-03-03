package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"unknwon.dev/i18n"
)

var (
	store  *i18n.Store
	locale string
	once   sync.Once
)

func Init() *i18n.Store {
	once.Do(func() {
		s := i18n.NewStore()
		locale = DetectLocale()

		dir := findLocaleFile()

		_, err := s.AddLocale("en-US", "English", dir)
		if err != nil {
			panic("failed to load locale: " + err.Error())
		}

		store = s
	})
	return store
}

func findLocaleFile() string {
	paths := []string{
		"i18n/locales/locale_en-US.ini",
		"../i18n/locales/locale_en-US.ini",
		"../../i18n/locales/locale_en-US.ini",
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		paths = append(paths,
			filepath.Join(exeDir, "i18n/locales/locale_en-US.ini"),
			filepath.Join(exeDir, "../i18n/locales/locale_en-US.ini"),
		)
	}

	for _, p := range paths {
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	return paths[0]
}

func DetectLocale() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		return "en-US"
	}
	return normalizeLocale(lang)
}

func normalizeLocale(l string) string {
	l = strings.TrimSpace(l)
	l = strings.ReplaceAll(l, "_", "-")
	l = strings.Split(l, ".")[0]
	if l == "" {
		return "en-US"
	}
	return l
}

func CurrentLocale() string {
	if locale == "" {
		Init()
	}
	return locale
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
