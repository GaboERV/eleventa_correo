package mailer

import (
	"fmt"
	"time"

	"github.com/GaboERV/reporteador-email/internal/models"
	"gopkg.in/gomail.v2"
)

// SendExcel emails the given excel file to the target email.
func SendExcel(cfg *models.Config, excelPath string, windowStart, windowEnd time.Time) error {
	m := gomail.NewMessage()
	
	// We assume Gmail for now
	m.SetHeader("From", cfg.SmtpUser)
	m.SetHeader("To", cfg.TargetEmail)
	
	subject := fmt.Sprintf("Reporte PDV - %s (%s)", cfg.BranchName, windowStart.Format("02/Jan"))
	m.SetHeader("Subject", subject)
	
	body := fmt.Sprintf("Hola,\n\nAdjunto se encuentra el reporte de ventas de la sucursal %s.\n\nPeriodo: %s al %s\n\nSaludos,\nReporteador Automático",
		cfg.BranchName,
		windowStart.Format("2006-01-02"),
		windowEnd.Format("2006-01-02"),
	)
	m.SetBody("text/plain", body)
	m.Attach(excelPath)

	d := gomail.NewDialer("smtp.gmail.com", 587, cfg.SmtpUser, cfg.SmtpPass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
