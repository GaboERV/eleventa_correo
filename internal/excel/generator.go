package excel

import (
	"fmt"
	"sort"
	"time"

	"github.com/GaboERV/reporteador-email/internal/models"
	"github.com/xuri/excelize/v2"
)

// Generate creates an Excel file based on Propuesta A:
// - Sheet 1: Resumen General (KPIs + Totales por Día + Totales por Departamento)
// - Sheet 2..N: One sheet per Date, containing turnos with detailed department sales and totals.
func Generate(report *models.Report, cfg *models.Config, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Setup initial sheet
	firstSheet := "Resumen"
	f.SetSheetName("Sheet1", firstSheet)

	// Styles
	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "1F497D"},
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "1F497D"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Size: 10, Color: "595959"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2F5597"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	subHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1F497D"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	// Format currency helper (Excel number format code)
	currencyFmt := "\"$\"#,##0.00"

	totalRowStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Color: "000000"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"F2F2F2"}, Pattern: 1},
		Border:       []excelize.Border{{Type: "top", Style: 1, Color: "000000"}, {Type: "bottom", Style: 2, Color: "000000"}},
		CustomNumFmt: &currencyFmt,
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &currencyFmt,
	})
	boldCurrencyStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true},
		CustomNumFmt: &currencyFmt,
	})

	// Calculate overall metrics
	var totalVentasGlobal int64
	var totalGananciasGlobal int64
	var totalTurnosGlobal int
	deptoGlobal := make(map[string]struct {
		Ventas    int64
		Ganancias int64
	})

	type DaySummary struct {
		Fecha     string
		Turnos    int
		Ventas    int64
		Ganancias int64
	}
	var daySummaries []DaySummary

	for _, day := range report.Days {
		var dayVentas int64
		var dayGanancias int64
		totalTurnosGlobal += len(day.Turnos)

		for _, turno := range day.Turnos {
			for _, v := range turno.VentasPorDepartamento {
				dayVentas += v.VentasCentavos
				dayGanancias += v.GananciasCentavos

				curr := deptoGlobal[v.Departamento]
				curr.Ventas += v.VentasCentavos
				curr.Ganancias += v.GananciasCentavos
				deptoGlobal[v.Departamento] = curr
			}
		}

		totalVentasGlobal += dayVentas
		totalGananciasGlobal += dayGanancias

		daySummaries = append(daySummaries, DaySummary{
			Fecha:     day.Fecha,
			Turnos:    len(day.Turnos),
			Ventas:    dayVentas,
			Ganancias: dayGanancias,
		})
	}

	// ==========================================
	// 1. HOJA DE RESUMEN
	// ==========================================
	f.SetCellValue("Resumen", "A1", fmt.Sprintf("REPORTE DE VENTAS - SUCURSAL %s", cfg.BranchName))
	f.SetCellStyle("Resumen", "A1", "A1", titleStyle)

	f.SetCellValue("Resumen", "A2", fmt.Sprintf("Generado: %s", time.Now().Format("02/01/2006 15:04")))
	f.SetCellStyle("Resumen", "A2", "A2", subtitleStyle)

	// KPI Cards
	f.SetCellValue("Resumen", "A4", "TOTAL VENTAS")
	f.SetCellValue("Resumen", "B4", float64(totalVentasGlobal)/100.0)
	f.SetCellValue("Resumen", "D4", "TOTAL GANANCIAS")
	f.SetCellValue("Resumen", "E4", float64(totalGananciasGlobal)/100.0)
	f.SetCellValue("Resumen", "G4", "TOTAL TURNOS")
	f.SetCellValue("Resumen", "H4", totalTurnosGlobal)

	f.SetCellStyle("Resumen", "B4", "B4", boldCurrencyStyle)
	f.SetCellStyle("Resumen", "E4", "E4", boldCurrencyStyle)

	// Tabla 1: Resumen por Día
	f.SetCellValue("Resumen", "A7", "Resumen Diario de Ventas")
	f.SetCellStyle("Resumen", "A7", "A7", boldStyle)

	headersDia := []string{"Fecha", "No. Turnos", "Ventas Totales", "Ganancias Totales"}
	for i, h := range headersDia {
		cell, _ := excelize.CoordinatesToCellName(i+1, 8)
		f.SetCellValue("Resumen", cell, h)
		f.SetCellStyle("Resumen", cell, cell, headerStyle)
	}

	row := 9
	for _, ds := range daySummaries {
		f.SetCellValue("Resumen", fmt.Sprintf("A%d", row), ds.Fecha)
		f.SetCellValue("Resumen", fmt.Sprintf("B%d", row), ds.Turnos)
		f.SetCellValue("Resumen", fmt.Sprintf("C%d", row), float64(ds.Ventas)/100.0)
		f.SetCellValue("Resumen", fmt.Sprintf("D%d", row), float64(ds.Ganancias)/100.0)

		f.SetCellStyle("Resumen", fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row), currencyStyle)
		row++
	}

	// Total Fila Resumen Diario
	f.SetCellValue("Resumen", fmt.Sprintf("A%d", row), "TOTAL")
	f.SetCellValue("Resumen", fmt.Sprintf("B%d", row), totalTurnosGlobal)
	f.SetCellFormula("Resumen", fmt.Sprintf("C%d", row), fmt.Sprintf("SUM(C9:C%d)", row-1))
	f.SetCellFormula("Resumen", fmt.Sprintf("D%d", row), fmt.Sprintf("SUM(D9:D%d)", row-1))
	f.SetCellStyle("Resumen", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), totalRowStyle)

	// Tabla 2: Resumen por Departamento
	row += 3
	f.SetCellValue("Resumen", fmt.Sprintf("A%d", row), "Resumen Acumulado por Departamento")
	f.SetCellStyle("Resumen", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), boldStyle)
	row++

	headersDepto := []string{"Departamento", "Ventas Totales", "Ganancias Totales"}
	for i, h := range headersDepto {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue("Resumen", cell, h)
		f.SetCellStyle("Resumen", cell, cell, headerStyle)
	}
	row++
	startDeptoRow := row

	// Sort departments by name
	var deptNames []string
	for k := range deptoGlobal {
		deptNames = append(deptNames, k)
	}
	sort.Strings(deptNames)

	for _, dName := range deptNames {
		dData := deptoGlobal[dName]
		f.SetCellValue("Resumen", fmt.Sprintf("A%d", row), dName)
		f.SetCellValue("Resumen", fmt.Sprintf("B%d", row), float64(dData.Ventas)/100.0)
		f.SetCellValue("Resumen", fmt.Sprintf("C%d", row), float64(dData.Ganancias)/100.0)

		f.SetCellStyle("Resumen", fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), currencyStyle)
		row++
	}

	// Total Fila Departamentos
	f.SetCellValue("Resumen", fmt.Sprintf("A%d", row), "TOTAL")
	f.SetCellFormula("Resumen", fmt.Sprintf("B%d", row), fmt.Sprintf("SUM(B%d:B%d)", startDeptoRow, row-1))
	f.SetCellFormula("Resumen", fmt.Sprintf("C%d", row), fmt.Sprintf("SUM(C%d:C%d)", startDeptoRow, row-1))
	f.SetCellStyle("Resumen", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), totalRowStyle)

	f.SetColWidth("Resumen", "A", "A", 25)
	f.SetColWidth("Resumen", "B", "D", 18)

	// Custom number formatting for currency columns
	numFmtStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &currencyFmt})

	// ==========================================
	// 2. HOJAS POR DÍA
	// ==========================================
	for _, day := range report.Days {
		sheetName := day.Fecha
		f.NewSheet(sheetName)

		f.SetCellValue(sheetName, "A1", fmt.Sprintf("VENTAS DEL DÍA: %s", day.Fecha))
		f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

		curRow := 3

		var dayStartRow int
		var dayEndRow int
		var turnoTotalCellsVentas []string
		var turnoTotalCellsGanancias []string

		for _, turno := range day.Turnos {
			// Subheader Turno
			sospechosoStr := ""
			if turno.Sospechoso {
				sospechosoStr = " (¡TURNO SOSPECHOSO!)"
			}
			turnoHeader := fmt.Sprintf("Turno #%d - Cajero: %s %s", turno.TurnoID, turno.Cajero, sospechosoStr)
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), turnoHeader)
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("C%d", curRow), subHeaderStyle)
			curRow++

			// Shift details
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("Inicio: %s | Término: %s", turno.InicioEn, turno.TerminoEn))
			curRow++

			// Table headers
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "Departamento")
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", curRow), "Ventas ($)")
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", curRow), "Ganancias ($)")
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("C%d", curRow), headerStyle)
			curRow++

			startTurnoRow := curRow

			for _, dept := range turno.VentasPorDepartamento {
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), dept.Departamento)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", curRow), float64(dept.VentasCentavos)/100.0)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", curRow), float64(dept.GananciasCentavos)/100.0)

				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", curRow), fmt.Sprintf("C%d", curRow), currencyStyle)
				curRow++
			}

			// Subtotal Turno
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("Total Turno #%d", turno.TurnoID))
			f.SetCellFormula(sheetName, fmt.Sprintf("B%d", curRow), fmt.Sprintf("SUM(B%d:B%d)", startTurnoRow, curRow-1))
			f.SetCellFormula(sheetName, fmt.Sprintf("C%d", curRow), fmt.Sprintf("SUM(C%d:C%d)", startTurnoRow, curRow-1))
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("C%d", curRow), totalRowStyle)

			turnoTotalCellsVentas = append(turnoTotalCellsVentas, fmt.Sprintf("B%d", curRow))
			turnoTotalCellsGanancias = append(turnoTotalCellsGanancias, fmt.Sprintf("C%d", curRow))

			curRow += 2 // Blank lines between turnos
		}

		// Total del Día
		if len(turnoTotalCellsVentas) > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "TOTAL DEL DÍA")
			
			// Join formula for total day
			var formulaVentas string
			var formulaGanancias string
			for i := range turnoTotalCellsVentas {
				if i > 0 {
					formulaVentas += "+"
					formulaGanancias += "+"
				}
				formulaVentas += turnoTotalCellsVentas[i]
				formulaGanancias += turnoTotalCellsGanancias[i]
			}
			f.SetCellFormula(sheetName, fmt.Sprintf("B%d", curRow), formulaVentas)
			f.SetCellFormula(sheetName, fmt.Sprintf("C%d", curRow), formulaGanancias)
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("C%d", curRow), totalRowStyle)
		}

		f.SetColWidth(sheetName, "A", "A", 30)
		f.SetColWidth(sheetName, "B", "C", 20)
		_ = dayStartRow
		_ = dayEndRow
		_ = numFmtStyle
	}

	// Save the file
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}
	return nil
}
