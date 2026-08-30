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

const (
	defaultDllDir    = `C:\Program Files (x86)\AbarrotesPDV`
	defaultOriginFDB = `C:\Program Files (x86)\AbarrotesPDV\db\PDVDATA.FDB`
	defaultTempFDB   = `C:\Users\Public\PDVDATA_SNAP_WORKER.FDB`
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
	
	cfg, err := models.LoadConfig(installDir)
	if err != nil {
		log.Fatalf("Error cargando config: %v", err)
	}

	if cfg.BranchName == "" || cfg.SmtpUser == "" || cfg.TargetEmail == "" {
		log.Fatalf("Configuración incompleta, abortando.")
	}

	// Calculate date range: since it's typically run on Monday, let's export the previous 7 days
	// Or dynamically: from last run date to today.
	// We'll use the last_run state to know the last time we reported.
	lastRun, err := models.LoadLastRun(installDir)
	var start time.Time
	if err == nil && lastRun.LastSuccessfulReportEnd != "" {
		start, _ = time.Parse(time.RFC3339, lastRun.LastSuccessfulReportEnd)
	} else {
		// Default to 7 days ago if no last run
		start = time.Now().AddDate(0, 0, -7)
	}

	// Truncate start to beginning of day
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := time.Now() // up to right now

	log.Printf("Extrayendo desde %s hasta %s", start.Format("2006-01-02"), end.Format("2006-01-02"))

	if err := extractor.CopyFDB(defaultOriginFDB, defaultTempFDB); err != nil {
		log.Fatalf("Error copiando DB: %v", err)
	}
	defer extractor.CleanupFDB(defaultTempFDB)

	dummyCfg := &models.Config{BranchName: cfg.BranchName}
	report, err := extractor.Extract(defaultDllDir, defaultTempFDB, start, end, dummyCfg)
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

	excelPath := filepath.Join(installDir, fmt.Sprintf("Reporte_%s.xlsx", time.Now().Format("20060102")))
	if err := excel.Generate(report, cfg, excelPath); err != nil {
		log.Fatalf("Error generando excel: %v", err)
	}

	if err := mailer.SendExcel(cfg, excelPath, start, end); err != nil {
		log.Fatalf("Error enviando correo: %v", err)
	}

	log.Println("Reporte enviado con éxito.")
	
	// Update last run
	if lastRun == nil {
		lastRun = &models.LastRun{}
	}
	lastRun.LastSuccessfulReportEnd = end.Format(time.RFC3339)
	models.SaveLastRun(installDir, lastRun)
	
	log.Println("Ejecución finalizada correctamente.")
}
