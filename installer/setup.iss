[Setup]
AppName=Reporteador PDV por Correo
AppVersion=2.0.0
DefaultDirName={pf}\ReporteadorPDVEmail
DefaultGroupName=Reporteador PDV
OutputBaseFilename=InstalarReporteadorCorreo
Compression=lzma
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64
PrivilegesRequired=admin

[Files]
Source: "..\release\reporteador-gui.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\worker.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Reporteador PDV Correo"; Filename: "{app}\reporteador-gui.exe"
Name: "{commondesktop}\Reporteador PDV"; Filename: "{app}\reporteador-gui.exe"

[Run]
; Tarea programada: Lunes a las 08:00 AM (El usuario puede cambiar esto en la GUI, pero configuramos un default)
Filename: "schtasks.exe"; Parameters: "/create /tn ""ReporteadorPDVEmail"" /tr ""'{app}\worker.exe'"" /sc weekly /d MON /st 08:00 /ru SYSTEM /f"; Flags: runhidden
; Lanzar la GUI para que el usuario configure las credenciales al terminar
Filename: "{app}\reporteador-gui.exe"; Flags: nowait postinstall; Description: "Configurar credenciales y enviar primer reporte"

[UninstallRun]
Filename: "schtasks.exe"; Parameters: "/delete /tn ""ReporteadorPDVEmail"" /f"; Flags: runhidden
