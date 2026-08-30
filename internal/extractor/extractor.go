package extractor

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/GaboERV/reporteador-email/internal/firebird"
	"github.com/GaboERV/reporteador-email/internal/models"
)

// The SQL query that extracts closed shifts with their sales by department.
// CRITICAL: This is a SELECT-only query. We NEVER modify the Firebird database.
const extractionSQL = `SELECT
  t.ID AS turno_id,
  u.ID AS cajero_id,
  u.NOMBRE_COMPLETO AS cajero,
  t.INICIO_EN,
  t.TERMINO_EN,
  t.SOSPECHOSO,
  d.ID AS departamento_id,
  d.NOMBRE AS departamento,
  cvd.ACUMULADO_VENTAS,
  cvd.ACUMULADO_GANANCIAS
FROM TURNOS t
JOIN USUARIOS u ON u.ID = t.ID_CAJERO
JOIN CORTE_VENTAS_POR_DEPTO cvd ON cvd.ID_TURNO = t.ID
JOIN DEPARTAMENTOS d ON d.ID = cvd.ID_DEPARTAMENTO
WHERE t.TERMINO_EN IS NOT NULL
  AND t.INICIO_EN >= ?
  AND t.INICIO_EN < ?
ORDER BY t.INICIO_EN ASC, t.ID ASC`

// Extract connects to a copy of the Firebird database, runs the extraction query,
// and groups the flat rows into the nested Report structure.
func Extract(dllDir, dbPath string, windowStart, windowEnd time.Time, cfg *models.Config) (*models.Report, error) {
	// Load fbclient.dll
	fb, err := firebird.LoadFBClient(dllDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load fbclient.dll: %w", err)
	}

	// Connect to the COPY of the database (never the original)
	conn, err := fb.Connect(dbPath, "SYSDBA", "masterkey")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Firebird: %w", err)
	}
	defer conn.Close()

	// Begin read-only transaction
	tx, err := conn.BeginTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Always rollback - we never write anything

	// Execute the extraction query with window parameters
	rs, err := conn.Query(tx, extractionSQL, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute extraction query: %w", err)
	}
	defer rs.Close()

	// Load timezone for parsing
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		loc = time.FixedZone("CST", -6*3600)
	}

	// Group flat rows into nested structure:
	// map[fecha] -> map[turnoID] -> *turnoBuilder
	type turnoBuilder struct {
		turno models.Turno
		deptos []models.VentaDepartamento
	}

	dayMap := make(map[string]map[int]*turnoBuilder) // fecha -> turnoID -> builder
	dayOrder := make([]string, 0)

	for rs.Next() {
		// Scan each column by position (matching the SELECT order)
		turnoID, _ := rs.ScanInt(0)
		cajeroID, _ := rs.ScanInt(1)
		cajero, _ := rs.ScanString(2)
		inicioEn, _ := rs.ScanTime(3, loc)
		terminoEn, _ := rs.ScanTime(4, loc)
		sospechoso, _ := rs.ScanBool(5)
		deptoID, _ := rs.ScanInt(6)
		deptoNombre, _ := rs.ScanString(7)
		ventasCentavos, _ := rs.ScanCentavos(8)
		gananciasCentavos, _ := rs.ScanCentavos(9)

		// Group by date of inicio_en
		fecha := inicioEn.Format("2006-01-02")

		if _, exists := dayMap[fecha]; !exists {
			dayMap[fecha] = make(map[int]*turnoBuilder)
			dayOrder = append(dayOrder, fecha)
		}

		turnos := dayMap[fecha]
		tb, exists := turnos[turnoID]
		if !exists {
			tb = &turnoBuilder{
				turno: models.Turno{
					TurnoID:    turnoID,
					CajeroID:   cajeroID,
					Cajero:     cajero,
					InicioEn:   inicioEn.Format(time.RFC3339),
					TerminoEn:  terminoEn.Format(time.RFC3339),
					Sospechoso: sospechoso,
				},
			}
			turnos[turnoID] = tb
		}

		// Add department sales to this turno
		tb.deptos = append(tb.deptos, models.VentaDepartamento{
			DepartamentoID:    deptoID,
			Departamento:      deptoNombre,
			VentasCentavos:    ventasCentavos,
			GananciasCentavos: gananciasCentavos,
		})
	}

	// Build the final nested structure
	sort.Strings(dayOrder)

	days := make([]models.Day, 0, len(dayOrder))
	for _, fecha := range dayOrder {
		turnoMap := dayMap[fecha]

		// Collect and sort turnos by ID
		turnoIDs := make([]int, 0, len(turnoMap))
		for id := range turnoMap {
			turnoIDs = append(turnoIDs, id)
		}
		sort.Ints(turnoIDs)

		turnos := make([]models.Turno, 0, len(turnoIDs))
		for _, id := range turnoIDs {
			tb := turnoMap[id]
			tb.turno.VentasPorDepartamento = tb.deptos
			turnos = append(turnos, tb.turno)
		}

		days = append(days, models.Day{
			Fecha:  fecha,
			Turnos: turnos,
		})
	}

	// Get hostname
	hostname, _ := os.Hostname()

	report := &models.Report{
		SchemaVersion: "1.0",
		BranchID:      cfg.BranchName,
		Hostname:      hostname,
		ClientVersion: "email-v1",
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Days:          days,
	}

	return report, nil
}
