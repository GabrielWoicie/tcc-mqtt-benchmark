param(
    [Parameter(Mandatory=$true)][string]$Broker,
    [Parameter(Mandatory=$true)][ValidateSet("emqx","mosquitto")][string]$BrokerName,
    [string]$Exe = ".\mqttbench.exe"
)

$ErrorActionPreference = "Stop"

foreach ($qos in 0,1,2) {
    & $Exe message `
        --broker $Broker `
        --broker-name $BrokerName `
        --clients 100 `
        --qos $qos `
        --rate 1 `
        --warmup 10s `
        --duration 60s `
        --drain 5s `
        --payload-bytes 256 `
        --connect-rate 100

    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    Start-Sleep -Seconds 20
}
