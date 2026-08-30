package excel

import (
	"fmt"
	"time"

	"github.com/GaboERV/reporteador-email/internal/models"
	"github.com/xuri/excelize/v2"
)

// Generate creates an Excel file for the given report and returns the file path.
func Generate(report *models.Report, cfg *models.Config, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// 1. Rename default sheet to Resumen
	f.SetSheetName("Sheet1", "Resumen")
	f.NewSheet("Turnos")
	f.NewSheet("Departamentos")

	// Calculate totals
	var totalVentas int64
	var totalGanancias int64
	var totalTurnos int
	deptoMap := make(map[string]int64)

	for _, day := range report.Days {
		totalTurnos += len(day.Turnos)
		for _, turno := range day.Turnos {
			for _, v := range turno.VentasPorDepartamento {
				totalVentas += v.VentasCentavos
				totalGanancias += v.GananciasCentavos
				deptoMap[v.Departamento] += v.VentasCentavos
			}
		}
	}

	// Styles
	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 164, // Built-in currency format or similar
	})

	// === RESUMEN SHEET ===
	f.SetCellValue("Resumen", "A1", "Reporte de Ventas")
	f.SetCellStyle("Resumen", "A1", "A1", boldStyle)
	f.SetCellValue("Resumen", "A2", "Sucursal:")
	f.SetCellValue("Resumen", "B2", cfg.BranchName)
	f.SetCellValue("Resumen", "A3", "Fecha de Generación:")
	f.SetCellValue("Resumen", "B3", time.Now().Format("2006-01-02 15:04:05"))

	f.SetCellValue("Resumen", "A5", "Total Ventas:")
	f.SetCellValue("Resumen", "B5", float64(totalVentas)/100.0)
	f.SetCellStyle("Resumen", "B5", "B5", currencyStyle)

	f.SetCellValue("Resumen", "A6", "Total Ganancias:")
	f.SetCellValue("Resumen", "B6", float64(totalGanancias)/100.0)
	f.SetCellStyle("Resumen", "B6", "B6", currencyStyle)

	f.SetCellValue("Resumen", "A7", "Total Turnos:")
	f.SetCellValue("Resumen", "B7", totalTurnos)

	// === TURNOS SHEET ===
	turnosHeaders := []string{"Turno ID", "Cajero", "Inicio", "Fin", "Sospechoso", "Ventas Turno", "Ganancias Turno"}
	for i, h := range turnosHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Turnos", cell, h)
		f.SetCellStyle("Turnos", cell, cell, headerStyle)
	}

	row := 2
	for _, day := range report.Days {
		for _, turno := range day.Turnos {
			var tVentas, tGanancias int64
			for _, v := range turno.VentasPorDepartamento {
				tVentas += v.VentasCentavos
				tGanancias += v.GananciasCentavos
			}
			f.SetCellValue("Turnos", fmt.Sprintf("A%d", row), turno.TurnoID)
			f.SetCellValue("Turnos", fmt.Sprintf("B%d", row), turno.Cajero)
			f.SetCellValue("Turnos", fmt.Sprintf("C%d", row), turno.InicioEn)
			f.SetCellValue("Turnos", fmt.Sprintf("D%d", row), turno.TerminoEn)
			
			sospechoso := "No"
			if turno.Sospechoso {
				sospechoso = "Sí"
			}
			f.SetCellValue("Turnos", fmt.Sprintf("E%d", row), sospechoso)
			
			f.SetCellValue("Turnos", fmt.Sprintf("F%d", row), float64(tVentas)/100.0)
			f.SetCellStyle("Turnos", fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), currencyStyle)
			
			f.SetCellValue("Turnos", fmt.Sprintf("G%d", row), float64(tGanancias)/100.0)
			f.SetCellStyle("Turnos", fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), currencyStyle)
			
			row++
		}
	}
	f.SetColWidth("Turnos", "A", "G", 20)

	// === DEPARTAMENTOS SHEET ===
	deptoHeaders := []string{"Departamento", "Ventas Totales"}
	for i, h := range deptoHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Departamentos", cell, h)
		f.SetCellStyle("Departamentos", cell, cell, headerStyle)
	}

	dRow := 2
	for depto, ventas := range deptoMap {
		f.SetCellValue("Departamentos", fmt.Sprintf("A%d", dRow), depto)
		f.SetCellValue("Departamentos", fmt.Sprintf("B%d", dRow), float64(ventas)/100.0)
		f.SetCellStyle("Departamentos", fmt.Sprintf("B%d", dRow), fmt.Sprintf("B%d", dRow), currencyStyle)
		dRow++
	}
	f.SetColWidth("Departamentos", "A", "B", 25)

	// Save the file
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}
	return nil
}
