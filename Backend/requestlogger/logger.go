package requestlogger

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type requestLog struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Status    int    `json:"status"`
	Latency   string `json:"latency"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	Bytes     int    `json:"bytes"`
	Error     string `json:"error"`
}

// New devuelve el middleware que escribe una linea JSON por request terminado.
//
// Sustituye al middleware `logger` de Fiber a proposito: su campo `Format`
// interpola ${path} y ${error} sin escapar, asi que una comilla o un salto de
// linea en cualquiera de los dos rompe el JSON y pierde esa linea el
// agregador. Aca se serializa con encoding/json, que escapa siempre.
//
// Los campos comunes con el SaaS (ts, level, service) son los que permiten
// juntar los logs de los dos servicios en una sola linea de tiempo.
func New(output io.Writer) fiber.Handler {
	if output == nil {
		output = os.Stdout
	}
	return func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		entry := requestLog{
			Timestamp: started.UTC().Format(time.RFC3339Nano),
			Level:     "info",
			Service:   "backend",
			Status:    c.Response().StatusCode(),
			Latency:   time.Since(started).String(),
			Method:    c.Method(),
			Path:      c.OriginalURL(),
			IP:        c.IP(),
			Bytes:     len(c.Response().Body()),
		}
		if err != nil {
			entry.Error = err.Error()
			if entry.Status < fiber.StatusBadRequest {
				entry.Status = fiber.StatusInternalServerError
			}
		}
		line, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			line = []byte(`{"level":"error","service":"backend","msg":"failed to encode request log"}`)
		}
		line = append(line, '\n')
		_, _ = output.Write(line)
		return err
	}
}
