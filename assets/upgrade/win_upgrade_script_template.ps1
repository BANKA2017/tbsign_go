$ErrorActionPreference = "Stop"
Start-Sleep -Seconds 1
$exePath = "%s"
$newExe = "%s"
$psScriptPath = $MyInvocation.MyCommand.Definition

do {
    Start-Sleep -Milliseconds 500
} while (
    (Test-Path $exePath) -and
    (Get-Process | Where-Object { $_.Path -eq $exePath } | Measure-Object).Count -gt 0
)

Move-Item -Force -Path $newExe -Destination $exePath

Remove-Item -Force $psScriptPath
