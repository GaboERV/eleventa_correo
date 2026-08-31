package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/GaboERV/reporteador-email/internal/excel"
	"github.com/GaboERV/reporteador-email/internal/extractor"
	"github.com/GaboERV/reporteador-email/internal/mailer"
	"github.com/GaboERV/reporteador-email/internal/models"
)

func main() {
	dir, _ := os.Executable()
	installDir := filepath.Dir(dir)

	logFile, err := os.OpenFile(filepath.Join(installDir, "worker.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("========================================")
	log.Println("Ejecución Automática (Worker) Iniciada")
	
	cfg, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("Error cargando config: %v", err)
	}

	if cfg.BranchName == "" || cfg.SmtpUser == "" || cfg.TargetEmail == "" {
		log.Fatalf("Configuración incompleta, abortando.")
	}

	// Fallbacks if advanced config is empty
	if cfg.DllDir == "" {
		cfg.DllDir = `C:\Program Files (x86)\AbarrotesPDV`
	}
	if cfg.OriginFDB == "" {
		cfg.OriginFDB = `C:\Program Files (x86)\AbarrotesPDV\db\PDVDATA.FDB`
	}
	if cfg.TempFDB == "" {
		cfg.TempFDB = `C:\Users\Public\PDVDATA_SNAP_MAIL.FDB`
	}

	lastRun, err := models.LoadLastRun()
	var start time.Time
	if err == nil && lastRun.LastSuccessfulReportEnd != "" {
		start, _ = time.Parse(time.RFC3339, lastRun.LastSuccessfulReportEnd)
	} else {
		start = time.Now().AddDate(0, 0, -7)
	}

	// Truncate start to beginning of day
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := time.Now()

	log.Printf("Extrayendo desde %s hasta %s", start.Format("2006-01-02"), end.Format("2006-01-02"))

	if err := extractor.CopyFDB(cfg.OriginFDB, cfg.TempFDB); err != nil {
		log.Fatalf("Error copiando DB: %v", err)
	}
	defer extractor.CleanupFDB(cfg.TempFDB)

	report, err := extractor.Extract(cfg.DllDir, cfg.TempFDB, start, end, cfg)
	if err != nil {
		log.Fatalf("Error extrayendo: %v", err)
	}

	turnoCount := 0
	for _, day := range report.Days {
		turnoCount += len(day.Turnos)
	}
	if turnoCount == 0 {
		log.Println("Sin turnos cerrados, no se enviará reporte.")
		return
	}

	excelPath := filepath.Join(`C:\ProgramData\ReporteadorPDVEmail`, fmt.Sprintf("Reporte_%s.xlsx", time.Now().Format("20060102")))
	if err := excel.Generate(report, cfg, excelPath); err != nil {
		log.Fatalf("Error generando excel: %v", err)
	}

	if err := mailer.SendExcel(cfg, excelPath, start, end); err != nil {
		log.Fatalf("Error enviando correo: %v", err)
	}

	log.Println("Reporte enviado con éxito.")
	
	if lastRun == nil {
		lastRun = &models.LastRun{}
	}
	lastRun.LastSuccessfulReportEnd = end.Format(time.RFC3339)
	models.SaveLastRun(lastRun)
	
	log.Println("Ejecución finalizada correctamente.")
}
