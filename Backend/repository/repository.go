// Package repository centraliza el acceso a los datos del sistema.
// Actualmente usa datos estáticos; está preparado para migrar a persistencia relacional.
package repository

import "multicatalogo-backend/models"

// ValidateCredentials comprueba las credenciales y devuelve el rol del usuario.
func ValidateCredentials(email, password string) (string, bool) {
	switch {
	case email == "admin@upse.edu.ec" && password == "123456":
		return "admin", true
	case email == "cliente@upse.edu.ec" && password == "123456":
		return "cliente", true
	default:
		return "", false
	}
}

// productosData almacena el catálogo de productos predefinidos.
var productosData = []models.Producto{
	{
		ID:          1,
		Nombre:      "Serum Revitalizante",
		Descripcion: "Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.",
		Precio:      45.0,
		Categoria:   "Serum",
		Img:         "https://picsum.photos/seed/serum/600",
		Galeria: []string{
			"https://picsum.photos/seed/serum1/600",
			"https://picsum.photos/seed/serum2/600",
			"https://picsum.photos/seed/serum3/600",
		},
	},
	{
		ID:          2,
		Nombre:      "Crema Hidratante Pro",
		Descripcion: "Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.",
		Precio:      32.5,
		Categoria:   "Crema",
		Img:         "https://picsum.photos/seed/crema/600",
		Galeria: []string{
			"https://picsum.photos/seed/crema1/600",
			"https://picsum.photos/seed/crema2/600",
			"https://picsum.photos/seed/crema3/600",
		},
	},
	{
		ID:          3,
		Nombre:      "Tónico Purificante",
		Descripcion: "Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.",
		Precio:      28.0,
		Categoria:   "Tónico",
		Img:         "https://picsum.photos/seed/tonico/600",
		Galeria: []string{
			"https://picsum.photos/seed/tonico1/600",
			"https://picsum.photos/seed/tonico2/600",
			"https://picsum.photos/seed/tonico3/600",
		},
	},
	{
		ID:          4,
		Nombre:      "Mascarilla Nocturna",
		Descripcion: "Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.",
		Precio:      50.0,
		Categoria:   "Mascarilla",
		Img:         "https://picsum.photos/seed/mascarilla/600",
		Galeria: []string{
			"https://picsum.photos/seed/mascarilla1/600",
			"https://picsum.photos/seed/mascarilla1/600",
			"https://picsum.photos/seed/mascarilla2/600",
			"https://picsum.photos/seed/mascarilla3/600",
		},
	},
	{
		ID:          5,
		Nombre:      "Serum Anti-Edad Retinol",
		Descripcion: "Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.",
		Precio:      58.0,
		Categoria:   "Serum",
		Img:         "https://picsum.photos/seed/retinol/600",
		Galeria: []string{
			"https://picsum.photos/seed/retinol1/600",
			"https://picsum.photos/seed/retinol2/600",
			"https://picsum.photos/seed/retinol3/600",
		},
	},
	{
		ID:          6,
		Nombre:      "Crema Contorno de Ojos",
		Descripcion: "Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.",
		Precio:      26.0,
		Categoria:   "Crema",
		Img:         "https://picsum.photos/seed/ojos/600",
		Galeria: []string{
			"https://picsum.photos/seed/ojos1/600",
			"https://picsum.photos/seed/ojos2/600",
			"https://picsum.photos/seed/ojos3/600",
		},
	},
	{
		ID:          7,
		Nombre:      "Tónico Exfoliante AHA-BHA",
		Descripcion: "Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.",
		Precio:      34.0,
		Categoria:   "Tónico",
		Img:         "https://picsum.photos/seed/exfoliante/600",
		Galeria: []string{
			"https://picsum.photos/seed/exfoliante1/600",
			"https://picsum.photos/seed/exfoliante2/600",
			"https://picsum.photos/seed/exfoliante3/600",
		},
	},
	{
		ID:          8,
		Nombre:      "Kit Rutina Completa",
		Descripcion: "Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.",
		Precio:      129.0,
		Categoria:   "Kit",
		Img:         "https://picsum.photos/seed/kit/600",
		Galeria: []string{
			"https://picsum.photos/seed/kit1/600",
			"https://picsum.photos/seed/kit2/600",
			"https://picsum.photos/seed/kit3/600",
		},
	},
}

// GetProductos devuelve el catálogo completo de productos.
func GetProductos() []models.Producto {
	return productosData
}

// GetProductoByID busca un producto por su identificador.
func GetProductoByID(id int) (models.Producto, bool) {
	for _, p := range productosData {
		if p.ID == id {
			return p, true
		}
	}
	return models.Producto{}, false
}

// redData contiene el árbol jerárquico de referidos multinivel.
var redData = models.Referido{
	ID:     0,
	Nombre: "Tú",
	Nivel:  0,
	Ventas: 2400,
	Hijos: []models.Referido{
		{
			ID:     1,
			Nombre: "Ana García",
			Nivel:  1,
			Ventas: 1200,
			Hijos: []models.Referido{
				{
					ID:     4,
					Nombre: "Carlos Ruiz",
					Nivel:  2,
					Ventas: 500,
					Hijos: []models.Referido{
						{ID: 7, Nombre: "Diana Paz", Nivel: 3, Ventas: 300},
					},
				},
				{ID: 5, Nombre: "Sofía León", Nivel: 2, Ventas: 430},
			},
		},
		{
			ID:     2,
			Nombre: "Luis Poveda",
			Nivel:  1,
			Ventas: 850,
			Hijos: []models.Referido{
				{ID: 6, Nombre: "Marco Díaz", Nivel: 2, Ventas: 380},
			},
		},
		{ID: 3, Nombre: "Marta Sánchez", Nivel: 1, Ventas: 430},
	},
}

// GetRed devuelve la estructura jerárquica de la red de referidos.
func GetRed() models.Referido {
	return redData
}
