package lang

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/dogsays/mo/cfgmgr"
	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/ut2"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// 升级版的语言包。支持复数。使用gotemplate

/*
[PersonCats]
description = "The number of cats a person has"
one = "{{.Name}} has {{.Count}} cat."
other = "{{.Name}} has {{.Count}} cats."

	lang.Translate("zh-CN", "PersonCats", map[string]any{
			"Name": "Nick",
			"Count": 2,
		}, 2)
*/
var bundle = i18n.NewBundle(language.English)

func Init(name string, cm *cfgmgr.ConfigManager) {
	cm.WatchAndLoad(name, loadLang)
}

func loadLang(buf []byte) error {

	reader := csv.NewReader(bytes.NewReader(buf))
	all, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return errors.New("empty")
	}

	header := all[0]
	if len(header) == 0 {
		return errors.New("empty header")
	}
	body := all[1:]

	files := make([]*i18n.MessageFile, len(header))
	for i, v := range header {
		tag, err := language.Parse(v)
		if err != nil {
			return err
		}

		files[i] = &i18n.MessageFile{Tag: tag}
	}

	for _, row := range body {
		id := row[0]
		key := "other"
		if idx := strings.IndexByte(id, '@'); idx != -1 {
			sp := id[idx+1:]
			switch sp {
			case "zero", "one", "two", "few", "many", "other":
				key = sp
				id = id[:idx]
			}
		}

		for i := 1; i < len(row); i++ {
			if row[i] != "" {
				mp := map[string]any{}
				mp["id"] = id
				mp[key] = row[i]

				file := files[i]

				msg, _ := i18n.NewMessage(mp)

				file.Messages = append(file.Messages, msg)
			}
		}
	}

	newB := i18n.NewBundle(language.English)

	for _, f := range files {
		err := newB.AddMessages(f.Tag, f.Messages...)
		if err != nil {
			logger.Info("多语言加载错误", f.Tag, err)
		}
	}

	bundle = newB
	cache.Clear()

	return err
}

var cache = ut2.NewSyncMap[string, *i18n.Localizer]()

func getLocalizer(language string) *i18n.Localizer {
	l, ok := cache.Load(language)
	if ok {
		return l
	}

	l = i18n.NewLocalizer(bundle, language)
	cache.Store(language, l)

	return l
}

func localize(language string, cfg *i18n.LocalizeConfig) string {
	lc := getLocalizer(language)

	str, err := lc.Localize(cfg)
	if err != nil {
		logger.Info("多语言错误", language, cfg.MessageID, err)
		return cfg.MessageID
	}

	return str
}

// 仅获取翻译文本
func Get(language string, id string) string {
	return localize(language, &i18n.LocalizeConfig{MessageID: id})
}

// 带参数的翻译
func Translate(language string, id string, data any, plural any) string {
	return localize(language, &i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
		PluralCount:  plural,
	})
}

// 配置里面通过 {{.arg1}} {{.arg2}} 配置
func Getf(language string, id string, args ...any) string {

	argMap := map[string]any{}
	for i, v := range args {
		argMap[fmt.Sprintf("arg%d", i+1)] = v
	}

	return Translate(language, id, argMap, 0)
}

// 错误
func Err(language string, id string) error {
	return errors.New(Get(language, id))
}
func Errf(language string, id string, args ...any) error {
	return errors.New(Getf(language, id, args...))
}
