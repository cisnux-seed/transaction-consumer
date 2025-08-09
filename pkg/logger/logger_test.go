package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger should not return nil")
	}

	// Test that the returned logger implements the Logger interface
	var _ Logger = logger
}

func TestNewLogger_AllMethods(t *testing.T) {
	// Test that NewLogger creates a logger that has all required methods
	logger := NewLogger()

	// Test that all methods exist and don't panic
	logger.Debug("test debug", "key", "value")
	logger.Info("test info", "key", "value")
	logger.Warn("test warn", "key", "value")
	logger.Error("test error", "key", "value")

	// Note: We can't test Fatal here as it would exit the test
}

func TestLogger_Debug(t *testing.T) {
	// Capture output
	var buf bytes.Buffer

	// Create logger with custom handler to capture output
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Actually call the Debug method on our logger instance
	logger.Debug("test debug message", "key", "value")

	// Check output contains expected elements
	output := buf.String()
	if !strings.Contains(output, "test debug message") {
		t.Error("Debug output should contain the message")
	}
	if !strings.Contains(output, "DEBUG") {
		t.Error("Debug output should contain DEBUG level")
	}

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Debug output should be valid JSON: %v", err)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Actually call the Info method on our logger instance
	logger.Info("test info message", "user", "john", "id", 123)

	output := buf.String()
	if !strings.Contains(output, "test info message") {
		t.Error("Info output should contain the message")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Info output should contain INFO level")
	}

	// Verify JSON structure
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Info output should be valid JSON: %v", err)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Actually call the Warn method on our logger instance
	logger.Warn("test warning message", "warning", true)

	output := buf.String()
	if !strings.Contains(output, "test warning message") {
		t.Error("Warn output should contain the message")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Warn output should contain WARN level")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Warn output should be valid JSON: %v", err)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Actually call the Error method on our logger instance
	logger.Error("test error message", "error", "something went wrong")

	output := buf.String()
	if !strings.Contains(output, "test error message") {
		t.Error("Error output should contain the message")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Error output should contain ERROR level")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Error output should be valid JSON: %v", err)
	}
}

func TestLogger_Fatal_Complete(t *testing.T) {
	// Check if this is the subprocess that should call Fatal
	if os.Getenv("GO_TEST_FATAL") == "1" {
		// This code runs in the subprocess and will actually call os.Exit(1)
		logger := NewLogger()
		logger.Fatal("fatal test message", "test", true)
		// This line should never be reached due to os.Exit(1)
		panic("Fatal did not exit")
	}

	// This is the main test process
	// Launch a subprocess to test the Fatal method
	cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal_Complete$")
	cmd.Env = append(os.Environ(), "GO_TEST_FATAL=1")

	// Capture the output
	output, err := cmd.CombinedOutput()

	// The subprocess should exit with code 1
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else {
		t.Errorf("Expected subprocess to exit with code 1, got error: %v", err)
	}

	// Check that the fatal message was logged before exit
	outputStr := string(output)
	if !strings.Contains(outputStr, "fatal test message") {
		t.Errorf("Expected fatal message in output, got: %s", outputStr)
	}
}

func TestLogger_AllMethodsWithCapture(t *testing.T) {
	// Test all logger methods with output capture to ensure 100% coverage
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Test Debug method
	logger.Debug("debug test", "debug", true)

	// Test Info method
	logger.Info("info test", "info", true)

	// Test Warn method
	logger.Warn("warn test", "warn", true)

	// Test Error method
	logger.Error("error test", "error", true)

	// Verify all messages were logged
	output := buf.String()

	expectedMessages := []string{"debug test", "info test", "warn test", "error test"}
	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Output should contain '%s'", msg)
		}
	}

	for _, level := range expectedLevels {
		if !strings.Contains(output, level) {
			t.Errorf("Output should contain '%s' level", level)
		}
	}

	// Count the number of log entries (should be 4)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Errorf("Expected 4 log entries, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d should be valid JSON: %v", i+1, err)
		}
	}
}

func TestLogger_MethodForwarding(t *testing.T) {
	// Test specific method implementations to ensure they call the right slog methods
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Test that Debug method forwards to slog.Debug
	logger.Debug("debug message")
	output1 := buf.String()
	buf.Reset()

	// Test that Info method forwards to slog.Info
	logger.Info("info message")
	output2 := buf.String()
	buf.Reset()

	// Test that Warn method forwards to slog.Warn
	logger.Warn("warn message")
	output3 := buf.String()
	buf.Reset()

	// Test that Error method forwards to slog.Error
	logger.Error("error message")
	output4 := buf.String()

	// Verify each method produces the expected log level
	if !strings.Contains(output1, "DEBUG") {
		t.Error("Debug method should produce DEBUG level log")
	}
	if !strings.Contains(output2, "INFO") {
		t.Error("Info method should produce INFO level log")
	}
	if !strings.Contains(output3, "WARN") {
		t.Error("Warn method should produce WARN level log")
	}
	if !strings.Contains(output4, "ERROR") {
		t.Error("Error method should produce ERROR level log")
	}

	// Verify messages are present
	if !strings.Contains(output1, "debug message") {
		t.Error("Debug method should log the message")
	}
	if !strings.Contains(output2, "info message") {
		t.Error("Info method should log the message")
	}
	if !strings.Contains(output3, "warn message") {
		t.Error("Warn method should log the message")
	}
	if !strings.Contains(output4, "error message") {
		t.Error("Error method should log the message")
	}
}

func TestLogger_WithMultipleArgs(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	logger.Info("processing transaction",
		"transactionId", "trans-123",
		"userId", 456,
		"amount", 100.50,
		"status", "success",
	)

	output := buf.String()

	// Verify the log entry is valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	// Check that structured data is present
	if logEntry["transactionId"] != "trans-123" {
		t.Error("Log should contain transactionId field")
	}
	if logEntry["userId"] != float64(456) {
		t.Error("Log should contain userId field")
	}
	if logEntry["amount"] != 100.50 {
		t.Error("Log should contain amount field")
	}
	if logEntry["status"] != "success" {
		t.Error("Log should contain status field")
	}
}

func TestLogger_WithNoArgs(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	logger.Info("simple message")

	output := buf.String()
	if !strings.Contains(output, "simple message") {
		t.Error("Output should contain the message")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	logger.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	// Verify required fields exist
	if _, exists := logEntry["time"]; !exists {
		t.Error("Log entry should contain time field")
	}
	if _, exists := logEntry["level"]; !exists {
		t.Error("Log entry should contain level field")
	}
	if _, exists := logEntry["msg"]; !exists {
		t.Error("Log entry should contain msg field")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	// Create logger with INFO level (should filter out DEBUG)
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	// Debug should be filtered out
	logger.Debug("debug message")
	debugOutput := buf.String()

	// Info should appear
	logger.Info("info message")
	infoOutput := buf.String()

	// Debug message should not appear in output
	if strings.Contains(debugOutput, "debug message") {
		t.Error("Debug message should be filtered out when level is INFO")
	}

	// Info message should appear
	if !strings.Contains(infoOutput, "info message") {
		t.Error("Info message should appear when level is INFO")
	}
}

func TestLogger_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slog: slog.New(handler),
	}

	logger.Info("")

	output := buf.String()
	if output == "" {
		t.Error("Logger should produce output even for empty message")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Output should be valid JSON even for empty message: %v", err)
	}
}

func TestLogger_Interface(t *testing.T) {
	// Test that logger implements the Logger interface
	var _ Logger = &logger{}

	// Test that NewLogger returns something that implements Logger interface
	var _ Logger = NewLogger()
}
