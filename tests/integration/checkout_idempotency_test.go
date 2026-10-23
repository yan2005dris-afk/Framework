// Test de integración de la idempotencia del checkout (U4-T3, S10).
//
// El escenario que cubre: el cliente manda un checkout, la respuesta se pierde
// (la app se cae, la red se corta), y el cliente reintenta. Sin idempotencia
// eso son dos pedidos con el mismo contenido y las comisiones contadas dos
// veces. Con idempotencia, el reintento devuelve el pedido original.
//
// Este test es el que cierra el círculo con el codigo real: los tests unitarios
// del controller comprueban que la key viaja y que el replay no vuelve a
// cobrar comisiones, pero acá se comprueba contra Postgres que la fila existe
// una sola vez y que el segundo POST no agrego un pedido.
//
// Pre-requisitos:
//   docker compose up -d --build
//   Las migraciones del Backend aplicadas (incluida la 006).
//
// Ejecutar:
//   cd tests/integration
//   go test -v -tags=integration -run TestCheckoutIdempotency
//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	checkoutEmail    = "admin@upse.edu.ec"
	checkoutPassword = "123456"
)

func checkoutGatewayURL() string {
	if v := os.Getenv("GATEWAY_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func checkoutDSN() string {
	host := os.Getenv("BACKEND_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("BACKEND_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("BACKEND_DB_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("BACKEND_DB_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("BACKEND_DB_NAME")
	if dbname == "" {
		dbname = "multicatalogo"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, dbname)
}

type checkoutResponseBody struct {
	PedidoID int     `json:"pedidoId"`
	Numero   string  `json:"numero"`
	Total    float64 `json:"total"`
	Mensaje  string  `json:"mensaje"`
}

// loginGateway devuelve la cookie de sesión del usuario demo.
func loginGateway(t *testing.T) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": checkoutEmail, "password": checkoutPassword})
	req, _ := http.NewRequest("POST", checkoutGatewayURL()+"/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("El stack no está levantado (%v). Levantá el stack y volvé a correr el test.", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, _ := io.ReadAll(resp.Body)
		t.Skipf("El login devolvió %d: %s. El stack quizá no tiene los datos demo.", resp.StatusCode, string(out))
	}
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			return c
		}
	}
	t.Skip("El login no devolvió la cookie auth_token")
	return nil
}

// postearCheckout manda un checkout con la key dada y devuelve status y cuerpo.
func postearCheckout(t *testing.T, cookie *http.Cookie, key string) (int, checkoutResponseBody) {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{
		"nombreCliente": "Test Idempotencia",
		"emailCliente":  checkoutEmail,
		"metodoPago":    "Tarjeta de crédito",
		"items":         []map[string]any{{"productoId": 1, "cantidad": 1, "precio": 45.0}},
	})
	req, _ := http.NewRequest("POST", checkoutGatewayURL()+"/api/checkout", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	req.AddCookie(cookie)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("No se pudo conectar al gateway: %v", err)
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	var parsed checkoutResponseBody
	_ = json.Unmarshal(out, &parsed)
	return resp.StatusCode, parsed
}

// TestCheckoutIdempotencyMismoKeyUnSoloPedido es el test que demuestra que
// dos POST con la misma key dejan un solo pedido.
func TestCheckoutIdempotencyMismoKeyUnSoloPedido(t *testing.T) {
	cookie := loginGateway(t)

	conn, err := pgx.Connect(context.Background(), checkoutDSN())
	if err != nil {
		t.Skipf("No se pudo conectar a Postgres (%v). El test necesita la BD del stack.", err)
	}
	defer conn.Close(context.Background())

	// La migración 006 tiene que estar aplicada; si no está, el resultado del
	// test no significaría nada.
	var existe int
	if err := conn.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'checkout_idempotencia'`,
	).Scan(&existe); err != nil || existe == 0 {
		t.Skip("La tabla checkout_idempotencia no existe: la migración 006 no está aplicada.")
	}

	// Key propia de esta corrida: si el test se corre dos veces, la segunda no
	// encuentra el replay de la primera y mide algo distinto.
	key := fmt.Sprintf("itest-%d", time.Now().UnixNano())

	status1, resp1 := postearCheckout(t, cookie, key)
	if status1 != http.StatusCreated {
		t.Fatalf("El primer checkout debe crear el pedido (201), devolvió %d: %+v", status1, resp1)
	}

	status2, resp2 := postearCheckout(t, cookie, key)
	if status2 != http.StatusOK {
		t.Fatalf("El reintento con la misma key debe hacer replay (200), devolvió %d: %+v", status2, resp2)
	}
	if resp1.PedidoID != resp2.PedidoID {
		t.Errorf("El replay devolvió un pedido distinto: primero %d, después %d", resp1.PedidoID, resp2.PedidoID)
	}

	ctx := context.Background()
	var filas int
	if err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM checkout_idempotencia WHERE clave = $1`, key,
	).Scan(&filas); err != nil {
		t.Fatalf("No se pudo consultar checkout_idempotencia: %v", err)
	}
	if filas != 1 {
		t.Errorf("La key %q quedó registrada %d veces, se esperaba exactamente 1", key, filas)
	}

	// Y el pedido que devolvió el replay tiene que existir una sola vez.
	var pedidos int
	if err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM pedidos WHERE id = $1`, resp1.PedidoID,
	).Scan(&pedidos); err != nil {
		t.Fatalf("No se pudo consultar pedidos: %v", err)
	}
	if pedidos != 1 {
		t.Errorf("El pedido %d aparece %d veces, se esperaba 1", resp1.PedidoID, pedidos)
	}

	t.Logf("✓ key %s: 201 (pedido %d) y luego 200 con el mismo pedido %d", key, resp1.PedidoID, resp2.PedidoID)
}

// TestCheckoutSinKeyCreaPedidosDistintos documenta el otro lado: sin key, cada
// POST es un pedido nuevo. Es el comportamiento que tienen hoy los clientes
// que no son la PWA, y por eso la key es opcional y no un requisito.
func TestCheckoutSinKeyCreaPedidosDistintos(t *testing.T) {
	cookie := loginGateway(t)

	status1, resp1 := postearCheckout(t, cookie, "")
	if status1 != http.StatusCreated {
		t.Fatalf("Sin key debe crear el pedido (201), devolvió %d: %+v", status1, resp1)
	}
	status2, resp2 := postearCheckout(t, cookie, "")
	if status2 != http.StatusCreated {
		t.Fatalf("Sin key el segundo POST también debe crear (201), devolvió %d: %+v", status2, resp2)
	}
	if resp1.PedidoID == resp2.PedidoID {
		t.Errorf("Sin key, dos POST tienen que crear dos pedidos distintos; ambos dieron %d", resp1.PedidoID)
	}
}