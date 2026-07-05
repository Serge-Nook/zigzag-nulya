; Inno Setup Script for Топор (Topor)
; Массовый редактор метаданных и временных штампов файлов

[Setup]
AppName=Топор
AppVersion=1.0.0
AppPublisher=Горшков Сергей Владимирович
AppPublisherURL=https://nookbat.ru/
DefaultDirName={autopf}\Topor
DefaultGroupName=Топор
OutputBaseFilename=Topor-1.0.0-Setup
OutputDir=..\..\dist\windows
Compression=lzma2
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
LicenseFile=..\..\LICENSE
SetupIconFile=topor.ico
UninstallDisplayIcon={app}\topor.exe
WizardStyle=modern
PrivilegesRequired=admin

[Languages]
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"

[Tasks]
Name: "desktopicon"; Description: "Создать ярлык на рабочем столе"; GroupDescription: "Дополнительные значки:"
Name: "addtopath"; Description: "Добавить в PATH"; GroupDescription: "Системные настройки:"; Flags: unchecked

[Files]
Source: "..\..\dist\topor\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs

[Icons]
Name: "{group}\Топор"; Filename: "{app}\topor.exe"; Comment: "Массовый редактор метаданных"
Name: "{group}\Удалить Топор"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Топор"; Filename: "{app}\topor.exe"; Tasks: desktopicon

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; \
    Tasks: addtopath; Check: NeedsAddPath(ExpandConstant('{app}'))

[Run]
Filename: "{app}\topor.exe"; Description: "Запустить Топор"; Flags: nowait postinstall skipifsilent

[Code]
function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE,
    'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
