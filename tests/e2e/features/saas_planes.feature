# language: es
Característica: Gestión y visualización de planes y suscripciones SaaS

  Antecedentes:
    Dado que el usuario administrador tiene una sesión activa en la plataforma
    Y navega a la sección de "Planes SaaS"

  Escenario: Visualización correcta de precios y características de los planes
    Cuando la vista de planes termina de cargar
    Entonces se muestran los 3 planes de suscripción: "Free", "Pro" y "Enterprise"
    Y el plan "Free" muestra un precio de "$0.00/mes"
    Y el plan "Pro" muestra un precio de "$29.90/mes"
    Y el plan "Enterprise" muestra un precio de "$99.90/mes"
    Y cada plan lista sus características y límites correspondientes

  Escenario: Suscripción a un plan de mayor capacidad
    Dado que el plan actual es "Free"
    Cuando el usuario hace clic en "Seleccionar Plan" en la tarjeta "Pro"
    Entonces se muestra el indicador de carga
    Y tras la confirmación del servidor el plan "Pro" pasa a tener la etiqueta "Plan Activo"

  Escenario: Cancelación de la suscripción actual
    Dado que el usuario tiene un plan activo contratado
    Cuando presiona el botón "Cancelar Suscripción Actual"
    Entonces el estado de suscripción se actualiza y se habilita la contratación nuevamente
