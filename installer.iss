#ifndef MyAppVersion
#define MyAppVersion "1.0.0"
#endif

[Setup]
AppId={{D3E8B5A1-9F2C-4A8E-8B7C-1234567890AB}
AppName=p-manager
AppVersion={#MyAppVersion}
AppPublisher=stepan41k
AppPublisherURL=https://github.com/stepan41k/p-manager
AppSupportURL=https://github.com/stepan41k/p-manager/issues
AppUpdatesURL=https://github.com/stepan41k/p-manager/releases
DefaultDirName={localappdata}\Programs\p-manager
DefaultGroupName=p-manager
OutputDir=dist
OutputBaseFilename=p-manager-windows-amd64-setup
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=lowest
ArchitecturesInstallIn64BitMode=x64

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "p-manager.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion; DestName: "README.txt"
Source: "LICENSE"; DestDir: "{app}"; Flags: ignoreversion; DestName: "LICENSE.txt"

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

[Code]
function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
