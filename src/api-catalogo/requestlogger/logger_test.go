package requestlogger

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// El path lo elige el cliente: si el serializer no escapa, una comilla o un
// salto de linea inyectan JSON invalido y el agregador pierde la linea entera.
// Este test es el que falla con el `Format` del middleware logger de Fiber.
func TestJSONRequestLoggerEscapesRequestControlledValues(t *testing.T) {
	var output bytes.Buffer
	app := fiber.New()
	app.Use(New(&output))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	response, err := app.Test(httptest.NewRequest("GET", "/?value=%22quoted%22%0Aline", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}

	if got := bytes.Count(output.Bytes(), []byte("\n")); got != 1 {
		t.Fatalf("expected exactly one JSON line, got %d newline(s): %q", got, output.String())
	}

	line := bytes.TrimSpace(output.Bytes())
	var payload map[string]any
	if err := json.Unmarshal(line, &payload); err != nil {
		t.Fatalf("log is not valid JSON: %v; line=%q", err, line)
	}
	if payload["status"] != float64(fiber.StatusOK) {
		t.Errorf("status field = %#v, want a JSON number", payload["status"])
	}
	if payload["path"] != "/?value=%22quoted%22%0Aline" {
		t.Errorf("path field = %#v", payload["path"])
	}
	if payload["error"] != "" {
		t.Errorf("error field = %#v, want empty on a successful request", payload["error"])
	}
	if payload["service"] != "backend" || payload["level"] != "info" {
		t.Errorf("service/level fields = %#v / %#v", payload["service"], payload["level"])
	}
}

// El `ts` es el campo con el que el agregador alinea los logs del Backend y
// del SaaS, asi que el formato exacto (RFC3339 en UTC con sufijo Z) es
// contrato entre los dos servicios, no un detalle de implementacion.
func TestJSONRequestLoggerEmiteTimestampUTCConSufijoZ(t *testing.T) {
	var output bytes.Buffer
	app := fiber.New()
	app.Use(New(&output))
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("ok") })

	response, err := app.Test(httptest.NewRequest("GET", "/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &payload); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	ts, ok := payload["ts"].(string)
	if !ok {
		t.Fatalf("ts field = %#v, want string", payload["ts"])
	}
	if !strings.HasSuffix(ts, "Z") {
		t.Errorf("ts = %q, want UTC with the Z suffix the SaaS also emits", ts)
	}
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		t.Errorf("ts = %q is not RFC3339: %v", ts, err)
	}
}

// Un handler que devuelve error no se puede "arreglar" en el logger: el
// middleware propaga el error y el status real (400) tiene que llegar intacto
// al cliente.
func TestJSONRequestLoggerPreservaElErrorDeFiber(t *testing.T) {
	app := fiber.New()
	app.Use(New(io.Discard))
	app.Get("/error", func(_ *fiber.Ctx) error { return fiber.ErrBadRequest })

	response, err := app.Test(httptest.NewRequest("GET", "/error", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusBadRequest)
	}
}
