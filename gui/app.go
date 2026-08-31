package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/GaboERV/reporteador-email/internal/excel"
	"github.com/GaboERV/reporteador-email/internal/extractor"
	"github.com/GaboERV/reporteador-email/internal/mailer"
	"github.com/GaboERV/reporteador-email/internal/models"
)

// App struct
type App struct {
	ctx        context.Context
	installDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	dir, _ := os.Executable()
	installDir := filepath.Dir(dir)
	return &App{installDir: installDir}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// LoadConfig returns the current configuration
func (a *App) LoadConfig() (*models.Config, error) {
	cfg, err := models.LoadConfig()
	if err != nil {
		// Return empty config with default paths if not exists
		return &models.Config{
			DllDir:    `C:\Program Files (x86)\AbarrotesPDV`,
			OriginFDB: `C:\Program Files (x86)\AbarrotesPDV\db\PDVDATA.FDB`,
			TempFDB:   `C:\Users\Public\PDVDATA_SNAP_MAIL.FDB`,
			AutoDay:   "Monday",
			AutoTime:  "08:00",
		}, nil
	}
	// Fallbacks if they saved empty advanced paths before
	if cfg.DllDir == "" {
		cfg.DllDir = `C:\Program Files (x86)\AbarrotesPDV`
	}
	if cfg.OriginFDB == "" {
		cfg.OriginFDB = `C:\Program Files (x86)\AbarrotesPDV\db\PDVDATA.FDB`
	}
	if cfg.TempFDB == "" {
		cfg.TempFDB = `C:\Users\Public\PDVDATA_SNAP_MAIL.FDB`
	}
	return cfg, nil
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(cfg *models.Config) error {
	return models.SaveConfig(cfg)
}

// GenerateAndSend manually extracts data and emails it
func (a *App) GenerateAndSend(startDateStr, endDateStr string) error {
	cfg, err := a.LoadConfig()
	if err != nil || cfg.BranchName == "" {
		return fmt.Errorf("configuración inválida o faltante")
	}

	start, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf("fecha inicio inválida")
	}
	end, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return fmt.Errorf("fecha fin inválida")
	}
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second) // End of day

	if err := extractor.CopyFDB(cfg.OriginFDB, cfg.TempFDB); err != nil {
		return fmt.Errorf("error copiando base de datos: %w", err)
	}
	defer extractor.CleanupFDB(cfg.TempFDB)

	// Extract
	report, err := extractor.Extract(cfg.DllDir, cfg.TempFDB, start, end, cfg)
	if err != nil {
		return fmt.Errorf("error en extracción: %w", err)
	}

	turnoCount := 0
	for _, day := range report.Days {
		turnoCount += len(day.Turnos)
	}
	if turnoCount == 0 {
		return fmt.Errorf("no hay turnos en el rango seleccionado")
	}

	// Generate Excel
	excelPath := filepath.Join(`C:\ProgramData\ReporteadorPDVEmail`, "reporte.xlsx")
	if err := excel.Generate(report, cfg, excelPath); err != nil {
		return fmt.Errorf("error generando excel: %w", err)
	}

	// Send Email
	if err := mailer.SendExcel(cfg, excelPath, start, end); err != nil {
		return fmt.Errorf("error enviando correo: %w", err)
	}

	return nil
}
