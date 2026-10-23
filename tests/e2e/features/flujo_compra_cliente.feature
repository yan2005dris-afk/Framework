# language: es
Característica: Flujo de compra completo en MultiCatálogo
  Como cliente registrado en la tienda
  Quiero explorar el catálogo, agregar productos al carrito y finalizar la compra
  Para adquirir los productos de belleza y verificar mi pedido

  Antecedentes:
    Dado que existe el usuario cliente con credenciales válidas en la plataforma

  Escenario: Inicio de sesión del cliente con nombre real
    Dado que el cliente "Mariana Morales" navega a la página de login
    Cuando ingresa su correo "cliente@upse.edu.ec" y contraseña "123456"
    Y presiona el botón de iniciar sesión
    Entonces es redirigido a la tienda o catálogo de productos
    Y visualiza el catálogo de productos disponibles

  Escenario: Selección de productos y adición al carrito
    Dado que el cliente está en el catálogo de productos
    Cuando busca el producto "Serum Revitalizante"
    Y añade el producto "Serum Revitalizante" al carrito de compras
    Y busca el producto "Crema Hidratante Pro"
    Y añade el producto "Crema Hidratante Pro" al carrito de compras
    Entonces el carrito de compras refleja 2 productos seleccionados

  Escenario: Finalización de compra y generación de pedido
    Dado que el cliente tiene productos en su carrito
    Y procede a la pantalla de finalización de compra
    Cuando completa los datos de envío con:
      | campo     | valor                       |
      | nombre    | Mariana Morales             |
      | email     | mariana.morales@upse.edu.ec |
      | direccion | Av. 9 de Octubre y Malecón  |
      | ciudad    | Guayaquil                   |
      | telefono  | 0998765432                  |
    Y selecciona el método de pago "Transferencia bancaria"
    Y confirma el pedido
    Entonces visualiza la pantalla de confirmación con el mensaje de éxito
    Y puede consultar el pedido en su historial de compras
