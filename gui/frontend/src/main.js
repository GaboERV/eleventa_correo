import { LoadConfig, SaveConfig, GenerateAndSend } from '../wailsjs/go/main/App';

document.addEventListener("DOMContentLoaded", () => {
    // Set default dates (start of month to today)
    const today = new Date();
    document.getElementById('end-date').value = today.toISOString().split('T')[0];
    const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    document.getElementById('start-date').value = startOfMonth.toISOString().split('T')[0];

    // Load config
    loadConfig();
});

function showToast(message, type = 'success') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const textSpan = document.createElement('span');
    textSpan.textContent = message;
    
    const closeBtn = document.createElement('span');
    closeBtn.className = 'toast-close';
    closeBtn.textContent = '×';
    closeBtn.onclick = () => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    };
    
    toast.appendChild(textSpan);
    toast.appendChild(closeBtn);
    container.appendChild(toast);
    
    // Trigger reflow for transition
    toast.offsetHeight;
    toast.classList.add('show');
    
    // Auto remove after 4 seconds
    setTimeout(() => {
        if (toast.parentElement) {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }
    }, 4000);
}

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
            
            // Advanced
            document.getElementById('dll-dir').value = cfg.dll_dir || '';
            document.getElementById('origin-fdb').value = cfg.origin_fdb || '';
            document.getElementById('temp-fdb').value = cfg.temp_fdb || '';
        }
    } catch (e) {
        showToast("Error cargando configuración: " + e, 'error');
        console.error("Error loading config:", e);
    }
}

window.saveConfig = async function() {
    const btnSave = document.getElementById('btn-save');
    const btnSaveAdv = document.getElementById('btn-save-adv');
    btnSave.disabled = true;
    btnSaveAdv.disabled = true;

    const cfg = {
        branch_name: document.getElementById('branch-name').value,
        smtp_user: document.getElementById('smtp-user').value,
        smtp_pass: document.getElementById('smtp-pass').value,
        target_email: document.getElementById('target-email').value,
        auto_day: document.getElementById('auto-day').value,
        auto_time: document.getElementById('auto-time').value,
        
        dll_dir: document.getElementById('dll-dir').value,
        origin_fdb: document.getElementById('origin-fdb').value,
        temp_fdb: document.getElementById('temp-fdb').value,
    };

    try {
        await SaveConfig(cfg);
        showToast("Configuración guardada exitosamente.", "success");
    } catch (e) {
        showToast("Error guardando: " + e, "error");
    } finally {
        btnSave.disabled = false;
        btnSaveAdv.disabled = false;
    }
}

window.sendReport = async function() {
    const btn = document.getElementById('btn-send');
    const startDate = document.getElementById('start-date').value;
    const endDate = document.getElementById('end-date').value;

    if (!startDate || !endDate) {
        showToast("Seleccione fecha de inicio y fin.", "error");
        return;
    }

    btn.disabled = true;
    const originalText = btn.textContent;
    btn.textContent = "Generando y Enviando...";

    try {
        await GenerateAndSend(startDate, endDate);
        showToast("Reporte generado y enviado exitosamente por correo.", "success");
    } catch (e) {
        showToast("Error: " + e, "error");
    } finally {
        btn.disabled = false;
        btn.textContent = originalText;
    }
}
