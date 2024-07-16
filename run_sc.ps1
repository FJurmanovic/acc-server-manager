param (
    [string]$Action,
    [string]$ServiceName
)

if ($Action -eq "start") {
    sc.exe start $ServiceName
} elseif ($Action -eq "stop") {
    sc.exe stop $ServiceName
} elseif ($Action -eq "restart") {
    sc.exe stop $ServiceName
    sc.exe start $ServiceName
} else {
    Write-Error "Invalid action specified. Use 'start', 'stop', or 'restart'."
}