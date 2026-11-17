// Package helpers contiene utilidades internas para los controllers.
package helpers

// SanitizeErrorDetails aplica el principio de mínima divulgación de errores:
// en producción (APP_ENV distinto de "dev") se devuelve solo el mensaje, sin
// el detalle crudo que podría exponer nombres de tablas, queries SQL o
// información interna.
//
// Recibe appEnv y err, devuelve (mensajePublico, detalleSeguro).
// - mensajePublico siempre es el mensaje del error.
// - detalleSeguro vale err.Error() solo si appEnv == "dev", vacío en otro caso.
func SanitizeErrorDetails(appEnv string, err error) (string, string) {
	if err == nil {
		return "", ""
	}
	msg := err.Error()
	if appEnv == "dev" {
		return msg, msg
	}
	return msg, ""
}
