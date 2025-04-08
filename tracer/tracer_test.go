package tracer

import "testing"

func TestInitStdoutTracer(t *testing.T) {
	tests := []struct {
		name    string
		config  StdoutConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: StdoutConfig{
				ServiceName: "test-service",
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			config: StdoutConfig{
				ServiceName: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitStdoutTracer(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
