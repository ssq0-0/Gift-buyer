package sessions

import "github.com/gotd/td/telegram"

func (f *sessionManagerImpl) CreateDeviceConfig() telegram.DeviceConfig {
	config := telegram.DeviceConfig{}

	config.SetDefaults()

	if f.cfg.DeviceModel != "" {
		config.DeviceModel = f.cfg.DeviceModel
	} else {
		config.DeviceModel = "MacBook Pro M1 Pro"
	}

	if f.cfg.SystemVersion != "" {
		config.SystemVersion = f.cfg.SystemVersion
	} else {
		config.SystemVersion = "macOS 14.1"
	}

	if f.cfg.AppVersion != "" {
		config.AppVersion = f.cfg.AppVersion
	} else {
		config.AppVersion = "11.9 (272031) APP_STORE"
	}

	if f.cfg.SystemLangCode != "" {
		config.SystemLangCode = f.cfg.SystemLangCode
	} else {
		config.SystemLangCode = "en"
	}

	if f.cfg.LangCode != "" {
		config.LangCode = f.cfg.LangCode
	} else {
		config.LangCode = "en"
	}

	if f.cfg.LangPack != "" {
		config.LangPack = f.cfg.LangPack
	} else {
		config.LangPack = "macos"
	}

	return config
}
