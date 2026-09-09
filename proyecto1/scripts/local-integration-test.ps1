$ErrorActionPreference = "Stop"

$projectDir = Split-Path -Parent $PSScriptRoot
$tempDir = Join-Path $projectDir ".tmp"
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

Push-Location $projectDir
$processes = @()
$savedEnvironment = @{
    GOFLAGS = $env:GOFLAGS
    GOTOOLCHAIN = $env:GOTOOLCHAIN
    GOCACHE = $env:GOCACHE
    GOMODCACHE = $env:GOMODCACHE
    CARNET = $env:CARNET
    VM_NAME = $env:VM_NAME
    PORT = $env:PORT
    API1_URL = $env:API1_URL
    API2_URL = $env:API2_URL
    API3_URL = $env:API3_URL
}

try {
    $env:GOFLAGS = "-buildvcs=false"
    $env:GOTOOLCHAIN = "local"
    $env:GOCACHE = Join-Path $tempDir "gocache"
    $env:GOMODCACHE = Join-Path $tempDir "gomodcache"
    go test ./...
    go build -o (Join-Path $tempDir "api1.exe") ./api1
    go build -o (Join-Path $tempDir "api2.exe") ./api2
    go build -o (Join-Path $tempDir "api3.exe") ./api3

    $env:CARNET = "201801391"

    $env:VM_NAME = "VM1"; $env:PORT = "8081"
    $env:API2_URL = "http://127.0.0.1:8082"; $env:API3_URL = "http://127.0.0.1:8083"
    $processes += Start-Process -FilePath (Join-Path $tempDir "api1.exe") -WindowStyle Hidden -PassThru

    $env:VM_NAME = "VM1"; $env:PORT = "8082"
    $env:API1_URL = "http://127.0.0.1:8081"; $env:API3_URL = "http://127.0.0.1:8083"
    $processes += Start-Process -FilePath (Join-Path $tempDir "api2.exe") -WindowStyle Hidden -PassThru

    $env:VM_NAME = "VM2"; $env:PORT = "8083"
    $env:API1_URL = "http://127.0.0.1:8081"; $env:API2_URL = "http://127.0.0.1:8082"
    $processes += Start-Process -FilePath (Join-Path $tempDir "api3.exe") -WindowStyle Hidden -PassThru

    Start-Sleep -Seconds 1

    $checks = @(
        @{ Url = "http://127.0.0.1:8081/health"; API = "API1" },
        @{ Url = "http://127.0.0.1:8082/health"; API = "API2" },
        @{ Url = "http://127.0.0.1:8083/health"; API = "API3" }
    )
    foreach ($check in $checks) {
        $response = Invoke-RestMethod -Uri $check.Url
        if ($response.status -ne "UP" -or $response.carnet -ne "201801391") {
            throw "Contrato /health inválido para $($check.API)"
        }
        $response | ConvertTo-Json -Compress
    }

    $callUrls = @(
        "http://127.0.0.1:8081/api1/201801391/call-api2",
        "http://127.0.0.1:8081/api1/201801391/call-api3",
        "http://127.0.0.1:8082/api2/201801391/call-api1",
        "http://127.0.0.1:8082/api2/201801391/call-api3",
        "http://127.0.0.1:8083/api3/201801391/call-api1",
        "http://127.0.0.1:8083/api3/201801391/call-api2"
    )
    foreach ($url in $callUrls) {
        $response = Invoke-RestMethod -Uri $url
        if (-not $response.connection -or $response.carnet -ne "201801391") {
            throw "Comunicación cruzada inválida en $url"
        }
        $response | ConvertTo-Json -Compress
    }

    "Prueba local completa: 3 endpoints /health y 6 llamadas cruzadas correctas."
}
finally {
    foreach ($process in $processes) {
        if ($null -ne $process -and -not $process.HasExited) {
            Stop-Process -Id $process.Id -Force
        }
    }
    foreach ($key in $savedEnvironment.Keys) {
        if ($null -eq $savedEnvironment[$key]) {
            Remove-Item -Path "Env:$key" -ErrorAction SilentlyContinue
        }
        else {
            Set-Item -Path "Env:$key" -Value $savedEnvironment[$key]
        }
    }
    Pop-Location
}
