// manifest_windows.go - Install and Uninstall manifest file for Windows.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//go:build windows

package host

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func (h *Host) getTargetNames() ([]string, error) {
	targets := []string{}
	registryKey := `SOFTWARE\Google`
	var access uint32 = registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS
	regKey, err := registry.OpenKey(registry.CURRENT_USER, registryKey, access)
	defer func() {
		if err := regKey.Close(); err != nil {
		}
	}()

	keyNames, err := regKey.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	} else {
		for _, keyName := range keyNames {
			if strings.HasPrefix(keyName, "Chrome") {
				fullKey := registryKey + `\` + keyName
				targets = append(targets, fullKey)
			}
		}
	}
	return targets, nil
}

// Install creates native-messaging manifest file on appropriate location and
// add an entry in windows registry. It will return error when it come across
// one.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location
func (h *Host) Install() error {
	manifest, _ := json.MarshalIndent(h, "", "  ")
	registryNames, err := h.getTargetNames()
	if err != nil {
		return err
	}
	for _, keyName := range registryNames {
		registryName := keyName + `\NativeMessagingHosts\` + h.AppName
		targetName := filepath.Join(filepath.Dir(h.ExecName), h.AppName+".json")

		if err := ioutilWriteFile(targetName, manifest, 0644); err != nil {
			return err
		}

		// CreateKey creates a key named path under open key k. CreateKey returns the
		// new key and a boolean flag that reports whether the key already existed.
		key, _, err := registry.CreateKey(registry.CURRENT_USER, registryName, registry.SET_VALUE)
		if err != nil {
			log.Printf(`Error installing: HKCU\%s: %v`, registryName, err)
		} else {
			defer key.Close()

			if err := key.SetStringValue("", targetName); err != nil {
				log.Printf(`Error installing: HKCU\%s: %v`, registryName, err)
			} else {

				log.Printf(`Installed: HKCU\%s`, registryName)
			}
		}
	}
	return nil
}

// Uninstall removes entry from windows registry and removes native-messaging
// manifest file from installed location.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location
func (h *Host) Uninstall() {
	registryNames, err := h.getTargetNames()
	if err != nil {
		return
	}

	for _, keyName := range registryNames {
		registryName := keyName + `\NativeMessagingHosts\` + h.AppName

		targetName := filepath.Join(filepath.Dir(h.ExecName), h.AppName+".json")

		key, err := registry.OpenKey(registry.CURRENT_USER, registryName, registry.SET_VALUE)
		if err != nil {
			log.Printf("Error opening %s: %v", registryName, err)
		} else {
			defer key.Close()

			if err := key.DeleteValue(""); err != nil {
				// It might never have been installed.
				log.Print(err)
			}

			if err := os.Remove(targetName); err != nil {
				// It might never have been installed.
				log.Print(err)
			}

			if err := os.Remove(h.ExecName); err != nil {
				// It might be locked by current process.
				log.Print(err)
			}

			if err := os.Remove(h.ExecName + ".chk"); err != nil {
				// It might not exist.
				log.Print(err)
			}

			log.Printf(`Uninstalled: HKCU\%s`, registryName)
		}

	}
	// Exit gracefully.
	runtimeGoexit()
}
