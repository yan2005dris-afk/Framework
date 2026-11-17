// Package events provee publicación y consumo de eventos usando Valkey Streams.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"multicatalogo-backend/config"
)

const (
	// StreamPedidos es el nombre del Valkey Stream para eventos de pedidos.
	StreamPedidos = "stream:pedidos"

	// EventPedidoCreado indica que se completó un checkout.
	EventPedidoCreado = "pedido.creado"
)

// PedidoCreadoEvent representa el payload de un pedido completado en api-catalogo.
type PedidoCreadoEvent struct {
	PedidoID  int       `json:"pedido_id"`
	TenantID  int       `json:"tenant_id"`
	Cliente   string    `json:"cliente"`
	Cantidad  int       `json:"cantidad"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

// Publisher define el contrato para publicar eventos del dominio.
type Publisher interface {
	PublishPedidoCreado(ctx context.Context, event PedidoCreadoEvent) (string, error)
}

// EventPublisher gestiona la publicación de eventos en Valkey Streams.
type EventPublisher struct {
	client *redis.Client
}

// NewEventPublisher crea una instancia del publicador usando el cliente Singleton de Valkey.
func NewEventPublisher() (*EventPublisher, error) {
	client, err := config.GetValkeyClient()
	if err != nil {
		return nil, fmt.Errorf("no se pudo inicializar EventPublisher: %w", err)
	}
	return &EventPublisher{client: client}, nil
}

// PublishPedidoCreado publica el evento pedido.creado en el Valkey Stream.
func (p *EventPublisher) PublishPedidoCreado(ctx context.Context, event PedidoCreadoEvent) (string, error) {
	if p == nil || p.client == nil {
		return "", fmt.Errorf("EventPublisher o cliente Valkey es nil")
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("error al serializar evento pedido.creado: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: StreamPedidos,
		Values: map[string]interface{}{
			"event_type": EventPedidoCreado,
			"payload":    string(payload),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	id, err := p.client.XAdd(ctx, args).Result()
	if err != nil {
		log.Printf("[events] Error publicando evento en Valkey Stream %s: %v", StreamPedidos, err)
		return "", fmt.Errorf("fallo al publicar en stream %s: %w", StreamPedidos, err)
	}

	log.Printf("[events] Evento %s publicado exitosamente en Valkey Stream [%s] con ID: %s", EventPedidoCreado, StreamPedidos, id)
	return id, nil
}
