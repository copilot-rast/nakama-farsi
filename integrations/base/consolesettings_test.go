package main

import "testing"

func TestConsoleSettingsClean(t *testing.T) {
	v, err := consoleSettings{}.clean()
	if err != nil || v.Timezone != "Asia/Tehran" {
		t.Fatalf("empty zone must default to Asia/Tehran, got %q %v", v.Timezone, err)
	}
	v, err = consoleSettings{Timezone: " Europe/Berlin "}.clean()
	if err != nil || v.Timezone != "Europe/Berlin" {
		t.Fatalf("zone must be trimmed, got %q %v", v.Timezone, err)
	}
	for _, bad := range []string{"Mars/Base", "local", "Local"} {
		if _, err := (consoleSettings{Timezone: bad}).clean(); err == nil {
			t.Fatalf("%q must be rejected", bad)
		}
	}
	if defaultConsoleSettings().Timezone != defaultConsoleTimezone {
		t.Fatal("defaults must use the Tehran zone")
	}
}
