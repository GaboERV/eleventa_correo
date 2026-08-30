import { LoadConfig, SaveConfig, GenerateAndSend } from '../../wailsjs/go/main/App';

document.addEventListener("DOMContentLoaded", () => {
    // Set default dates (start of month to today)
    const today = new Date();
    document.getElementById('end-date').value = today.toISOString().split('T')[0];
    const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    document.getElementById('start-date').value = startOfMonth.toISOString().split('T')[0];

    // Load config
    loadConfig();
});

async function loadConfig() {
    try {
        const cfg = await LoadConfig();
        if (cfg) {
            document.getElementById('branch-name').value = cfg.branch_name || '';
            document.getElementById('smtp-user').value = cfg.smtp_user || '';
            document.getElementById('smtp-pass').value = cfg.smtp_pass || '';
            document.getElementById('target-email').value = cfg.target_email || '';
            document.getElementById('auto-day').value = cfg.auto_day || 'Monday';
            document.getElementById('auto-time').value = cfg.auto_time || '08:00';
        }
    } catch (e) {
        console.error("Error loading config:", e);
    }
}

window.saveConfig = async function() {
    const btn = document.getElementById('btn-save');
    const status = document.getElementById('config-status');
    btn.disabled = true;
    status.className = 'status';

    const cfg = {
        branch_name: document.getElementById('branch-name').value,
        smtp_user: document.getElementById('smtp-user').value,
        smtp_pass: document.getElementById('smtp-pass').value,
        target_email: document.getElementById('target-email').value,
        auto_day: document.getElementById('auto-day').value,
        auto_time: document.getElementById('auto-time').value,
    };

    try {
        await SaveConfig(cfg);
        status.textContent = "Configuración guardada exitosamente.";
        status.classList.add('success');
    } catch (e) {
        status.textContent = "Error: " + e;
        status.classList.add('error');
    } finally {
        btn.disabled = false;
        setTimeout(() => status.classList.remove('success', 'error'), 3000);
    }
}

window.sendReport = async function() {
    const btn = document.getElementById('btn-send');
    const status = document.getElementById('send-status');
    const startDate = document.getElementById('start-date').value;
    const endDate = document.getElementById('end-date').value;

    if (!startDate || !endDate) {
        status.textContent = "Seleccione fecha de inicio y fin.";
        status.classList.add('error');
        return;
    }

    btn.disabled = true;
    btn.textContent = "Generando y Enviando...";
    status.className = 'status';

    try {
        await GenerateAndSend(startDate, endDate);
        status.textContent = "Reporte enviado exitosamente por correo.";
        status.classList.add('success');
    } catch (e) {
        status.textContent = "Error: " + e;
        status.classList.add('error');
    } finally {
        btn.disabled = false;
        btn.textContent = "Generar y Enviar por Correo";
    }
}
