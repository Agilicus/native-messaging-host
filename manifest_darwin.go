// manifest_darwin.go - Install and Uninstall manifest file for OS X.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// getTargetName returns an absolute path to native messaging host manifest
// location for OS X.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location-nix
func (h *Host) getTargetNames() ([]string, error) {
	targets := []string{}
	if os.Getuid() != 0 {
		homeDir, _ := os.UserHomeDir()
		matches, err := filepath.Glob(homeDir + "/Library/Application Support/Google/Chrome/NativeMessagingHosts")
		if err != nil {
			return nil, err
		}
		for _, s := range matches {
			target := filepath.Join(s, h.AppName+".json")
			targets = append(targets, target)
		}
	}
	return targets, nil
}

// Install creates native-messaging manifest file on appropriate location. It
// will return error when it come across one.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location-nix
func (h *Host) Install() error {
	manifest, _ := json.MarshalIndent(h, "", "  ")
	targetNames, err := h.getTargetNames()
	if err != nil {
		return err
	}

	for _, targetName := range targetNames {
		if err := osMkdirAll(filepath.Dir(targetName), 0755); err != nil {
			return err
		}

		if err := ioutilWriteFile(targetName, manifest, 0644); err != nil {
			return err
		}

		log.Printf("Installed: %s", targetName)
	}
	return nil
}

// Uninstall removes native-messaging manifest file from installed location.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location-nix
func (h *Host) Uninstall() {
	targetNames, err := h.getTargetNames()

	if err != nil {
		return
	}

	for _, targetName := range targetNames {
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

		log.Printf("Uninstalled: %s", targetName)
	}

	// Exit gracefully.
	runtimeGoexit()
}
