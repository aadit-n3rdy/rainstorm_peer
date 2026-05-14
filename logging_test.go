package main

import "testing"

func TestConfigureLoggerRejectsInvalidLevel(t *testing.T) {
	err := ConfigureLogger(AppConfig{LogLevel: "verbose", LogFormat: "console"})
	if err == nil {
		t.Fatal("expected invalid log level to return an error")
	}
}

func TestConfigureLoggerRejectsInvalidFormat(t *testing.T) {
	err := ConfigureLogger(AppConfig{LogLevel: "info", LogFormat: "xml"})
	if err == nil {
		t.Fatal("expected invalid log format to return an error")
	}
}

func TestConfigureLoggerAcceptsJSONAndConsole(t *testing.T) {
	for _, format := range []string{"json", "console"} {
		if err := ConfigureLogger(AppConfig{LogLevel: "debug", LogFormat: format}); err != nil {
			t.Fatalf("ConfigureLogger(%q) returned error: %v", format, err)
		}
	}
}
