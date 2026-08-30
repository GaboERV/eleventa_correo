package models

// Report is the top-level structure sent from client to server
type Report struct {
	SchemaVersion string `json:"schema_version"`
	BranchID      string `json:"branch_id"`
	Hostname      string `json:"hostname"`
	ClientVersion string `json:"client_version"`
	GeneratedAt   string `json:"generated_at"`
	Days          []Day  `json:"dias"`
}

// Day groups turnos by their start date
type Day struct {
	Fecha  string  `json:"fecha"`
	Turnos []Turno `json:"turnos"`
}

// Turno represents a closed cashier shift
type Turno struct {
	TurnoID               int                 `json:"turno_id"`
	CajeroID              int                 `json:"cajero_id"`
	Cajero                string              `json:"cajero"`
	InicioEn              string              `json:"inicio_en"`
	TerminoEn             string              `json:"termino_en"`
	Sospechoso            bool                `json:"sospechoso"`
	VentasPorDepartamento []VentaDepartamento `json:"ventas_por_departamento"`
}

// VentaDepartamento holds sales and profit for a department within a turno.
// Amounts are always in centavos (integer) to avoid floating point errors.
type VentaDepartamento struct {
	DepartamentoID    int    `json:"departamento_id"`
	Departamento      string `json:"departamento"`
	VentasCentavos    int64  `json:"ventas_centavos"`
	GananciasCentavos int64  `json:"ganancias_centavos"`
}
