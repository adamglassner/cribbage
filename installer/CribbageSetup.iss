; Cribbage - standalone setup. Build: bash build.sh, then ISCC.exe /DGameVersion=<version> installer\CribbageSetup.iss
; (c) 2026 Ironite
#define GameName    "Cribbage"
#define Bundle      "Cribbage"
#define Exe         "Cribbage.exe"
#ifndef GameVersion
  #define GameVersion "1.0.0"
#endif
#define AppGuid     "{{B3DCB672-4A62-43CD-876B-B15AA970BC98}"
#define Art         "cribbage"
#define Tagline     "A lighthouse card game. Play the cottage regulars, share one PC, or play friends online."
#define SaveWhat    "Cribbage record and settings"
#include "GameSetup.iss"

[Run]
; Let Cribbage host online games without a firewall prompt.
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Cribbage"" dir=in action=allow program=""{app}\Games\{#Bundle}\{#Exe}"" enable=yes"; Flags: runhidden; StatusMsg: "Letting Cribbage host online games..."

[UninstallRun]
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Cribbage"""; Flags: runhidden; RunOnceId: "CribbageFirewall"
