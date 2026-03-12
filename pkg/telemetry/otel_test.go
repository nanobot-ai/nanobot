package telemetry

import "testing"

func TestExportEnabled(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")

	tests := []struct {
		name               string
		tracesExporter     string
		otlpEndpoint       string
		otlpTracesEndpoint string
		want               bool
	}{
		{
			name: "no env disables export",
			want: false,
		},
		{
			name:           "traces exporter enables export",
			tracesExporter: "otlp",
			want:           true,
		},
		{
			name:         "otlp endpoint enables export",
			otlpEndpoint: "http://collector:4318",
			want:         true,
		},
		{
			name:               "otlp traces endpoint enables export",
			otlpTracesEndpoint: "http://collector:4318/v1/traces",
			want:               true,
		},
		{
			name:           "none disables export",
			tracesExporter: "none",
			want:           false,
		},
		{
			name:           "none overrides otlp endpoint",
			tracesExporter: "none",
			otlpEndpoint:   "http://collector:4318",
			want:           false,
		},
		{
			name:               "none overrides otlp traces endpoint",
			tracesExporter:     "none",
			otlpTracesEndpoint: "http://collector:4318/v1/traces",
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_TRACES_EXPORTER", tt.tracesExporter)
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tt.otlpEndpoint)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", tt.otlpTracesEndpoint)

			if got := exportEnabled(); got != tt.want {
				t.Fatalf("exportEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
