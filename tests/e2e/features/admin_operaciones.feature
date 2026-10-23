# language: es
Característica: Administración y gestión de operaciones en MultiCatálogo
  Como usuario administrador del sistema
  Quiero supervisar el panel administrativo, gestionar el catálogo de productos, revisar los pedidos de clientes y consultar la red multinivel
  Para mantener la tienda operando y actualizada con información consistente

  Antecedentes:
    Dado que el administrador "Carlos Mendoza" ha iniciado sesión en el sistema

  Escenario: Monitoreo del Panel de Control (Dashboard)
    Dado que el administrador se encuentra en el panel de control principal
    Cuando la vista del dashboard termina de cargar sus métricas
    Entonces se visualizan las tarjetas resumen de ventas y operaciones
    Y se listan las tablas de los productos más vendidos y mejores referidos

  Escenario: Gestión del catálogo y creación de nuevo producto
    Dado que el administrador navega a la sección de "Administración de Productos"
    Cuando abre el formulario para crear un nuevo producto
    Y completa los datos del producto con:
      | campo       | valor                           |
      | nombre      | Crema Antiedad Rosa Mosqueta    |
      | descripcion | Regeneración celular intensiva  |
      | precio      | 38.50                           |
      | stock       | 50                              |
    Y guarda el producto
    Entonces el nuevo producto "Crema Antiedad Rosa Mosqueta" aparece listado en el inventario

  Escenario: Supervisión y actualización de pedidos de clientes
    Dado que el administrador navega a la sección de "Gestión de Pedidos"
    Cuando visualiza la lista de órdenes registradas
    Entonces puede consultar el detalle del pedido más reciente
    Y puede filtrar los pedidos según su estado de entrega

  Escenario: Visualización de la red de referidos y comisiones MLM
    Dado que el administrador ingresa a la sección "Mi Red"
    Cuando carga la estructura multinivel
    Entonces visualiza los distribuidores y referidos de su organización
    Y observa el desglose de comisiones acumuladas por nivel
